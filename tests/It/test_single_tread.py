import unittest
import utils
import time


class TestSum(unittest.TestCase):
    promotion_state = 'test'
    pending_state = 'pending'

    def setUp(self) -> None:
        utils.truncate()

    def add_x_jobs(self, num: int = 10, cmd: str = 'exit 0', ttr: str = '10000') -> list:
        for i in range(num):
            utils.query(f"INSERT INTO jobs (ttr, cmd, status) VALUES({ttr},'{cmd}', '{self.promotion_state}')")

        actual = []
        for row in utils.query(f"SELECT * from jobs WHERE status in ('{self.promotion_state}') ORDER BY id"):
            actual.append(row)

        utils.query(f"UPDATE jobs SET status='{self.pending_state}' WHERE status IN ( '{self.promotion_state}')")
        return actual

    def is_processed(self) -> bool:
        num = 0
        for row in utils.query(
                f"SELECT * from jobs WHERE status in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id"):
            num += 1
        return num < 1

    def test_success_jobs(self):
        actual = self.add_x_jobs(num=10, cmd='exit 0', ttr='1000')

        for i in range(5):
            if self.is_processed():
                break
            time.sleep(i)

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'SUCCESS')
        self.assertEqual(len(actual), len(curr))

    def test_failed_jobs(self):
        actual = self.add_x_jobs(num=10, cmd='exit 1', ttr='1000')

        for i in range(5):
            if self.is_processed():
                break
            time.sleep(i)

        curr = utils.query(
            f"SELECT * from jobs WHERE status not in ('{self.pending_state}', '{self.promotion_state}') ORDER BY id")
        for row in curr:
            self.assertEqual(row['status'], 'FAILED')
        self.assertEqual(len(actual), len(curr))


if __name__ == '__main__':
    unittest.main()
