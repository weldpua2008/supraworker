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
        return num < 1

    def is_not_running(self) -> bool:
        num = 0
        for row in utils.query(
                f"SELECT * from jobs WHERE status not in ('{self.running_state}') ORDER BY id"):
            num += 1
        return num > 0

    def wait_processed(self, num):
        for i in range(num):
            if self.is_processed():
                break
            time.sleep(i*2)
        time.sleep(2)

    @num_jobs(5)
    def test_success_jobs(self, num):
        actual = self.add_x_jobs(num=num, cmd='exit 0', ttr='1000')
        self.wait_processed(num=num)

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'SUCCESS')
        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(5)
    def test_failed_jobs(self, num):
        actual = self.add_x_jobs(num=num, cmd='exit 1', ttr='100000')
        self.wait_processed(num=num)

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'FAILED')
        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(5)
    def test_cancelled_jobs(self, num):
        actual = self.add_x_jobs(num=num, cmd='sleep 10000', ttr='1000000')
        self.wait_processed(num=num)

        time.sleep(3)
        utils.query(
            f"UPDATE jobs SET status='{self.cancelled_state}' WHERE status IN ('{self.running_state}')")

        for i in range(num):
            if self.is_processed():
                break
            time.sleep(i)
        time.sleep(5)

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'CANCELLED')
        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

    @num_jobs(2)
    def test_timeout_jobs(self, num):
        actual = self.add_x_jobs(num=num, cmd='sleep 10000', ttr='5')
        self.wait_processed(num=num)
        time.sleep(10)

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        time.sleep(4)
        for row in curr:
            self.assertEqual(row['status'], 'TIMEOUT')

        self.assertEqual(len(actual), len(curr))
        self.assertEqual(len(curr), num)

if __name__ == '__main__':
    unittest.main()
