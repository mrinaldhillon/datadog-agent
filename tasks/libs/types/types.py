import io
import subprocess
from collections import defaultdict
from enum import Enum

from gitlab.v4.objects import ProjectJob


class Test:
    PACKAGE_PREFIX = "github.com/DataDog/datadog-agent/"

    def __init__(self, owners, name, package):
        self.name = name
        self.package = self.__removeprefix(package)
        self.owners = self.__get_owners(owners)

    def __removeprefix(self, package):
        return package[len(self.PACKAGE_PREFIX) :]

    def __find_file(self):
        # Find the *_test.go file in the package folder that has the test
        try:
            output = subprocess.run(
                [f"grep -Rl --include=\"*_test.go\" '{self.name}' '{self.package}'"],
                shell=True,
                stdout=subprocess.PIPE,
            )
            return output.stdout.decode('utf-8').splitlines()[0]
        except Exception as e:
            print(f"Exception '{e}' while finding test {self.name} from package {self.package}.")
            print("Setting file to '.none' to notify Agent Developer Experience")
            return '.none'

    def __get_owners(self, OWNERS):
        owners = OWNERS.of(self.__find_file())
        return [name for (kind, name) in owners if kind == "TEAM"]

    @property
    def key(self):
        return (self.name, self.package)


class FailedJobType(Enum):
    JOB_FAILURE = 1
    INFRA_FAILURE = 2
    BRIDGE_FAILURE = 3


class FailedJobReason(Enum):
    RUNNER = 1
    FAILED_JOB_SCRIPT = 5
    GITLAB = 6
    EC2_SPOT = 8
    E2E_INFRA_FAILURE = 9
    FAILED_BRIDGE_JOB = 10

    @staticmethod
    def get_infra_failure_mapping():
        return {
            'runner_system_failure': FailedJobReason.RUNNER,
            'stuck_or_timeout_failure': FailedJobReason.GITLAB,
            'unknown_failure': FailedJobReason.GITLAB,
            'api_failure': FailedJobReason.GITLAB,
            'scheduler_failure': FailedJobReason.GITLAB,
            'stale_schedule': FailedJobReason.GITLAB,
            'data_integrity_failure': FailedJobReason.GITLAB,
        }

    @staticmethod
    def from_gitlab_job_failure_reason(failure_reason: str):
        return FailedJobReason.get_infra_failure_mapping().get(failure_reason, FailedJobReason.GITLAB)


class FailedJobs:
    def __init__(self):
        self.mandatory_job_failures = []
        self.optional_job_failures = []
        self.mandatory_infra_job_failures = []
        self.optional_infra_job_failures = []

    def add_failed_job(self, job: ProjectJob):
        if job.failure_type == FailedJobType.INFRA_FAILURE and job.allow_failure:
            self.optional_infra_job_failures.append(job)
        elif job.failure_type == FailedJobType.INFRA_FAILURE and not job.allow_failure:
            self.mandatory_infra_job_failures.append(job)
        elif job.allow_failure:
            self.optional_job_failures.append(job)
        else:
            self.mandatory_job_failures.append(job)

    def all_non_infra_failures(self):
        return self.mandatory_job_failures + self.optional_job_failures

    def all_mandatory_failures(self):
        return self.mandatory_job_failures + self.mandatory_infra_job_failures

    def all_failures(self):
        return (
            self.mandatory_job_failures
            + self.optional_job_failures
            + self.mandatory_infra_job_failures
            + self.optional_infra_job_failures
        )


class SlackMessage:
    JOBS_SECTION_HEADER = "Failed jobs:"
    INFRA_SECTION_HEADER = "Infrastructure failures:"
    TEST_SECTION_HEADER = "Failed tests:"
    MAX_JOBS_PER_TEST = 2

    def __init__(self, base: str = "", jobs: FailedJobs = None, skipped: list | None = None):
        jobs = jobs if jobs else FailedJobs()
        self.base_message = base
        self.failed_jobs = jobs
        self.failed_tests = defaultdict(list)
        self.coda = ""
        self.skipped_jobs = skipped or []

    def add_test_failure(self, test, job):
        self.failed_tests[test.key].append(job)

    def __render_jobs_section(self, header: str, jobs: list, buffer: io.StringIO):
        if not jobs:
            return

        print(header, file=buffer)

        jobs_per_stage = defaultdict(list)
        for job in jobs:
            jobs_per_stage[job.stage].append(job)

        for stage, jobs in jobs_per_stage.items():
            jobs_info = []
            for job in jobs:
                num_retries = len(job.retry_summary) - 1
                emoji = " :job-skipped-on-pr:" if job.name in self.skipped_jobs else ""
                job_info = f"<{job.web_url}|{job.name}>{emoji}"
                if num_retries > 0:
                    job_info += f" ({num_retries} retries)"

                jobs_info.append(job_info)

            print(
                f"- {', '.join(jobs_info)} (`{stage}` stage)",
                file=buffer,
            )

    def __render_tests_section(self, buffer):
        print(self.TEST_SECTION_HEADER, file=buffer)
        for (test_name, test_package), jobs in self.failed_tests.items():
            job_list = ", ".join(f"<{job.web_url}|{job.name}>" for job in jobs[: self.MAX_JOBS_PER_TEST])
            if len(jobs) > self.MAX_JOBS_PER_TEST:
                job_list += f" and {len(jobs) - self.MAX_JOBS_PER_TEST} more"
            print(f"- `{test_name}` from package `{test_package}` (in {job_list})", file=buffer)

    def __str__(self):
        buffer = io.StringIO()
        if self.base_message:
            print(self.base_message, file=buffer)
        self.__render_jobs_section(
            self.JOBS_SECTION_HEADER,
            self.failed_jobs.mandatory_job_failures,
            buffer,
        )
        self.__render_jobs_section(
            self.INFRA_SECTION_HEADER,
            self.failed_jobs.mandatory_infra_job_failures,
            buffer,
        )
        if self.failed_tests:
            self.__render_tests_section(buffer)
        if self.coda:
            print(self.coda, file=buffer)
        return buffer.getvalue()


class PermissionCheck(Enum):
    """
    Enum to have a choice of permissions as argument to the check-permissions task.
    """

    REPO = 'repo'
    TEAM = 'team'
