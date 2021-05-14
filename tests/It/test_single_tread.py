import unittest
import utils
import time
import functools
import inspect

def num_jobs(number):
    def actual_decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, number, **kwargs)

        return wrapper

    return actual_decorator


NUM_WORKERS = 40


def wait_for_status(status, job_id) -> str:
    finished = False
    for i in range(0, NUM_WORKERS + 300):
        if finished:
            break
        for row in utils.query(f"SELECT * from jobs WHERE id={job_id}"):
            if row and row['status'] in status and str(job_id) in str(row['id']):
                finished = True
                break
            time.sleep(min(i, 2))


def wait_all_jobs(status) -> None:

    curr = utils.query(
        f"SELECT * from jobs ORDER BY id")
    for row in utils.progressbar(curr):
        wait_for_status(job_id=row['id'], status=status)


class TestSum(unittest.TestCase):
    propagated_state = 'propagated'
    promotion_state = 'test'
    pending_state = 'PENDING'
    running_state = 'RUNNING'
    cancelled_state = 'cancel'

    def setUp(self) -> None:
        utils.truncate()

    def tearDown(self) -> None:
        utils.query(
            f"UPDATE jobs SET status='{self.cancelled_state}' WHERE status not IN ('{self.cancelled_state}') ")
        time.sleep(5)

    def add_x_jobs(self, num: int = 10, cmd: str = 'exit 0', ttr: str = '10000') -> list:
        for i in range(num):
            utils.query(f"INSERT INTO jobs (ttr, cmd, status) VALUES({ttr},'{cmd}', '{self.promotion_state}')")

        actual = []
        for row in utils.query(
                f"SELECT * from jobs WHERE status in ('{self.promotion_state}') ORDER BY id"):
            actual.append(row)

        utils.query(
            f"UPDATE jobs SET status='{self.pending_state}' WHERE status IN ('{self.promotion_state}')")
        return actual

    def num_processed(self) -> int:
        n = utils.query(
            f"select count(*) as n  from jobs WHERE status not in ('{self.pending_state}', '{self.running_state}','{self.promotion_state}')")
        num = 0
        if n:
            num = n[0]['n']
        return num

    def wait_all_jobs_and_add(self, status: str, num: int, cmd: str, ttr: str) -> list:
        print()
        print(f"Adding {num} jobs for {inspect.stack()[1][3]}")
        actual = self.add_x_jobs(num=num, cmd=cmd, ttr=ttr)
        wait_all_jobs(status=status)

        return actual

    @num_jobs(NUM_WORKERS)
    def test_success_jobs(self, num):
        actual = self.wait_all_jobs_and_add(status='SUCCESS', num=num, cmd='exit 0', ttr='1000')
        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'SUCCESS')
        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(NUM_WORKERS)
    def test_failed_jobs(self, num):
        actual = self.wait_all_jobs_and_add(status='FAILED', num=num, cmd='exit 1', ttr='10100')
        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'FAILED')
        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(NUM_WORKERS)
    def test_cancelled_jobs(self, num):
        actual = self.wait_all_jobs_and_add(status=self.running_state, num=num, cmd='sleep 10000', ttr='1000000')
        utils.query(
            f"UPDATE jobs SET status='{self.cancelled_state}' WHERE status IN ('{self.running_state}')")

        wait_all_jobs('CANCELLED')

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'CANCELLED')
        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(NUM_WORKERS)
    def test_timeout_jobs(self, num):
        actual = self.wait_all_jobs_and_add(status='TIMEOUT', num=num, cmd='sleep 10000', ttr='3')

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'TIMEOUT')

        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(NUM_WORKERS)
    def test_success_jobs_more_than_workers(self, n):
        num = n * 2
        actual = self.wait_all_jobs_and_add(status='SUCCESS', num=num, cmd='exit 0', ttr='1001')
        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'SUCCESS')
        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(NUM_WORKERS)
    def test_timeout_jobs_more_than_workers(self, n):
        num = n * 2
        actual = self.wait_all_jobs_and_add(status='TIMEOUT', num=num, cmd='sleep 10000', ttr='1')

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'TIMEOUT')

        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(NUM_WORKERS)
    def test_failed_jobs_more_than_workers(self, n):
        num = n * 2
        # num = 1
        actual = self.wait_all_jobs_and_add(status='FAILED', num=num, cmd='exit 1', ttr='10100')
        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'FAILED')

        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)


if __name__ == '__main__':
    unittest.main()
