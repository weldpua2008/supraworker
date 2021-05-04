import unittest
import utils
import time
import functools


def num_jobs(number):
    def actual_decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, number, **kwargs)

        return wrapper

    return actual_decorator


NUM_JOBS = 5


class TestSum(unittest.TestCase):
    propogated_state = 'propogated'
    promotion_state = 'test'
    pending_state = 'PENDING'
    running_state = 'RUNNING'
    cancelled_state = 'cancel'

    def setUp(self) -> None:
        utils.truncate()

    def tearDown(self) -> None:
        utils.query(
            f"UPDATE jobs SET status='{self.cancelled_state}' WHERE status not IN ('{self.cancelled_state}') ")
        time.sleep(3)

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

    def is_processed(self) -> bool:
        num = 0
        for row in utils.query(
                f"SELECT * from jobs WHERE status in ('{self.pending_state}', '{self.running_state}','{self.promotion_state}') ORDER BY id"):
            num += 1
        return not (num > 0)

    def is_not_running(self) -> bool:
        num = 0
        for row in utils.query(
                f"SELECT * from jobs WHERE status not in ('{self.running_state}') ORDER BY id"):
            num += 1
        return num > 0

    def num_processed(self) -> int:
        n = utils.query(
            f"select count(*) as n  from jobs WHERE status not in ('{self.pending_state}', '{self.running_state}','{self.promotion_state}')")
        num = 0
        if n:
            num = n[0]['n']
        return num

    def wait_processed(self, num):
        for i in range(0, num):
            if self.is_processed():
                break
            time.sleep(i)
        for i in range(0, num):
            if self.num_processed() > num:
                break
            time.sleep(i)
        time.sleep(2)

    def wait_for_status(self, status, id) -> str:
        print(f"Waiting '{id}' for {status}", end='')
        for i in range(0, 3000):
            for row in utils.query(f"SELECT * from jobs WHERE id={id}"):
                if row['status'] in status:
                    break
                print(f".{i}", end='')
                time.sleep(1)
        print("")

    def wait_all_jobs_and_add(self, status, num, cmd, ttr) -> list:
        print(f"adding {num} jobs")
        actual = self.add_x_jobs(num=num, cmd=cmd, ttr=ttr)
        self.wait_all_jobs(status=status)

        return actual

    def wait_all_jobs(self, status) -> None:

        curr = utils.query(
            f"SELECT * from jobs ORDER BY id")
        print("waiting...")
        for row in curr:
            self.wait_for_status(id=row['id'], status=status)


    # @num_jobs(NUM_JOBS)
    # def test_success_jobs(self, num):
    #     actual = self.wait_all_jobs_and_add(status='SUCCESS', num=num, cmd='exit 0', ttr='1000')
    #     curr = utils.query(
    #         f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
    #     for row in curr:
    #         self.assertEqual(row['status'], 'SUCCESS')
    #     self.assertEqual(len(actual), len(curr))
    #     self.assertEqual(len(curr), num)
    # #
    # @num_jobs(NUM_JOBS)
    # def test_failed_jobs(self, num):
    #     actual = self.wait_all_jobs_and_add(status='FAILED', num=num, cmd='exit 1', ttr='10100')
    #     curr = utils.query(
    #         f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
    #     for row in curr:
    #         self.assertEqual(row['status'], 'FAILED')
    #     self.assertEqual(len(actual), len(curr))
    #     self.assertEqual(len(curr), num)
    # #
    @num_jobs(NUM_JOBS)
    def test_cancelled_jobs(self, num):
        actual = self.wait_all_jobs_and_add(status=self.running_state, num=num, cmd='sleep 10000', ttr='1000000')
        utils.query(
            f"UPDATE jobs SET status='{self.cancelled_state}' WHERE status IN ('{self.running_state}')")

        self.wait_all_jobs('CANCELLED')
        
        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'CANCELLED')
        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)
    #
    # @num_jobs(NUM_JOBS)
    # def test_timeout_jobs(self, num):
    #     actual = self.add_x_jobs(num=num, cmd='sleep 10000', ttr='5')
    #     self.wait_processed(num=num)
    #     time.sleep(10)
    #
    #     curr = utils.query(
    #         f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
    #     time.sleep(4)
    #     for row in curr:
    #         self.assertEqual(row['status'], 'TIMEOUT')
    #
    #     self.assertEqual(len(actual), len(curr))
    #     self.assertEqual(len(curr), num)


if __name__ == '__main__':
    unittest.main()
