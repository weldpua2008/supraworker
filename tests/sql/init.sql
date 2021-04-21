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
INSERT INTO jobs (ttr) VALUES(11);
INSERT INTO jobs (ttr) VALUES(12);
INSERT INTO jobs (ttr) VALUES(13);
INSERT INTO jobs (ttr) VALUES(14);
INSERT INTO jobs (ttr) VALUES(15);
INSERT INTO jobs (ttr) VALUES(16);
INSERT INTO jobs (ttr) VALUES(17);
INSERT INTO jobs (ttr) VALUES(18);
INSERT INTO jobs (ttr) VALUES(19);
INSERT INTO jobs (ttr) VALUES(20);
INSERT INTO jobs (ttr) VALUES(21);
INSERT INTO jobs (ttr) VALUES(22);
INSERT INTO jobs (ttr) VALUES(23);
INSERT INTO jobs (ttr) VALUES(24);
INSERT INTO jobs (ttr) VALUES(25);
INSERT INTO jobs (ttr) VALUES(26);
INSERT INTO jobs (ttr) VALUES(27);
INSERT INTO jobs (ttr) VALUES(28);
INSERT INTO jobs (ttr) VALUES(29);
INSERT INTO jobs (ttr) VALUES(30);
INSERT INTO jobs (ttr) VALUES(31);
INSERT INTO jobs (ttr) VALUES(32);
INSERT INTO jobs (ttr) VALUES(33);
INSERT INTO jobs (ttr) VALUES(34);
INSERT INTO jobs (ttr) VALUES(35);
INSERT INTO jobs (ttr) VALUES(36);
INSERT INTO jobs (ttr) VALUES(37);
INSERT INTO jobs (ttr) VALUES(38);
INSERT INTO jobs (ttr) VALUES(39);
INSERT INTO jobs (ttr) VALUES(40);
INSERT INTO jobs (ttr) VALUES(41);
INSERT INTO jobs (ttr) VALUES(42);
INSERT INTO jobs (ttr) VALUES(43);
INSERT INTO jobs (ttr) VALUES(44);
INSERT INTO jobs (ttr) VALUES(45);
INSERT INTO jobs (ttr) VALUES(46);
INSERT INTO jobs (ttr) VALUES(47);
INSERT INTO jobs (ttr) VALUES(48);
INSERT INTO jobs (ttr) VALUES(49);
INSERT INTO jobs (ttr) VALUES(50);
INSERT INTO jobs (ttr) VALUES(51);
INSERT INTO jobs (ttr) VALUES(52);
INSERT INTO jobs (ttr) VALUES(53);
INSERT INTO jobs (ttr) VALUES(54);
INSERT INTO jobs (ttr) VALUES(55);
INSERT INTO jobs (ttr) VALUES(56);
INSERT INTO jobs (ttr) VALUES(57);
INSERT INTO jobs (ttr) VALUES(58);
INSERT INTO jobs (ttr) VALUES(59);
INSERT INTO jobs (ttr) VALUES(60);
INSERT INTO jobs (ttr) VALUES(61);
INSERT INTO jobs (ttr) VALUES(62);
INSERT INTO jobs (ttr) VALUES(63);
INSERT INTO jobs (ttr) VALUES(64);
INSERT INTO jobs (ttr) VALUES(65);
INSERT INTO jobs (ttr) VALUES(66);
INSERT INTO jobs (ttr) VALUES(67);
INSERT INTO jobs (ttr) VALUES(68);
INSERT INTO jobs (ttr) VALUES(69);
INSERT INTO jobs (ttr) VALUES(70);
INSERT INTO jobs (ttr) VALUES(71);
INSERT INTO jobs (ttr) VALUES(72);
INSERT INTO jobs (ttr) VALUES(73);
INSERT INTO jobs (ttr) VALUES(74);
INSERT INTO jobs (ttr) VALUES(75);
INSERT INTO jobs (ttr) VALUES(76);
INSERT INTO jobs (ttr) VALUES(77);
INSERT INTO jobs (ttr) VALUES(78);
INSERT INTO jobs (ttr) VALUES(79);
INSERT INTO jobs (ttr) VALUES(80);
INSERT INTO jobs (ttr) VALUES(81);
INSERT INTO jobs (ttr) VALUES(82);
INSERT INTO jobs (ttr) VALUES(83);
INSERT INTO jobs (ttr) VALUES(84);
INSERT INTO jobs (ttr) VALUES(85);
INSERT INTO jobs (ttr) VALUES(86);
INSERT INTO jobs (ttr) VALUES(87);
INSERT INTO jobs (ttr) VALUES(88);
INSERT INTO jobs (ttr) VALUES(89);
INSERT INTO jobs (ttr) VALUES(90);
INSERT INTO jobs (ttr) VALUES(91);
INSERT INTO jobs (ttr) VALUES(92);
INSERT INTO jobs (ttr) VALUES(93);
INSERT INTO jobs (ttr) VALUES(94);
INSERT INTO jobs (ttr) VALUES(95);
INSERT INTO jobs (ttr) VALUES(96);
INSERT INTO jobs (ttr) VALUES(97);
INSERT INTO jobs (ttr) VALUES(98);
INSERT INTO jobs (ttr) VALUES(99);


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
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
INSERT INTO jobs (ttr, cmd) VALUES(100,'sleep 100');
