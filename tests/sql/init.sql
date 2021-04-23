CREATE DATABASE IF NOT EXISTS dev;
USE dev;
CREATE TABLE IF NOT EXISTS jobs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    run_id VARCHAR(255)  DEFAULT 'run_id',
    extra_run_id VARCHAR(255)  DEFAULT 'extra_run_id',
    ttr INT DEFAULT 30,
    status VARCHAR(255)  DEFAULT 'PENDING',
    cmd VARCHAR(255)  DEFAULT 'sleep 100',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )  ENGINE=INNODB;
/*
generating regular jobs
for i in range(3,100):
    print(f"INSERT INTO jobs (ttr) VALUES({i});")

 */

INSERT INTO jobs (ttr) VALUES(1);
INSERT INTO jobs (ttr) VALUES(2);
INSERT INTO jobs (ttr) VALUES(3);
INSERT INTO jobs (ttr) VALUES(4);
INSERT INTO jobs (ttr) VALUES(5);
INSERT INTO jobs (ttr) VALUES(6);
INSERT INTO jobs (ttr) VALUES(7);
INSERT INTO jobs (ttr) VALUES(8);
INSERT INTO jobs (ttr) VALUES(9);
INSERT INTO jobs (ttr) VALUES(10);


/*
 generating jobs for cancellation

 for i in range(101,201):
    print(f"INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');")
 */
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
