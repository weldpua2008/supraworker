import mysql.connector
import config
import logging


def query(sql, params=()):
    mysql_config = {
        'user': config.config.get('MYSQL_DATABASE_USER'),
        'database': config.config.get('MYSQL_DATABASE_DB'),
        'password': config.config.get('MYSQL_DATABASE_PASSWORD'),
        'host': config.config.get('MYSQL_DATABASE_HOST'),
        'port': config.config.get('MYSQL_DATABASE_PORT')
    }

    cnx = mysql.connector.connect(**mysql_config)
    response = []
    cursor = cnx.cursor()
    try:
        cursor.execute(sql, params)
        if 'UPDATE' in sql or 'INSERT' in sql or 'DELETE' in sql or 'ALTER' in sql:
            cnx.commit()
        result = []
        columns = cursor.description
        for value in cursor.fetchall():
            tmp = {}
            for (index, column) in enumerate(value):
                tmp[columns[index][0]] = column
            result.append(tmp)
        response = result
    except mysql.connector.Error as error:
        logging.warning(f"Can't perform {sql} got {error}")
    return response


def truncate():
    query(f"TRUNCATE TABLE { config.config.get('MYSQL_DATABASE_TABLE')}")
    logging.info("Truncating table")
    query(f"ALTER TABLE { config.config.get('MYSQL_DATABASE_TABLE')}  AUTO_INCREMENT=0")
    logging.info("Reseting autoincrement")

    # query(
    #     f"UPDATE jobs SET status='propogated' WHERE id={elem['job_uid']} AND status IN ( 'pending','PENDING')")