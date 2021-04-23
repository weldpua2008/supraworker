import logging
from random import randint
import time
from werkzeug.exceptions import BadRequest
from flask import request
import traceback
import mysql.connector

from flask import Blueprint, current_app, make_response, Response
from flask_restx import Namespace, Model, Api, Resource, fields, reqparse

logger = logging.getLogger(__name__)
job_page = Blueprint('job_page', __name__)
jobs_api = Api(
    job_page,
    version='1.0',
    title='Jobs API',
    default='jobs',
    default_label='Jobs related operations',
    description='The Jobs API allows you to run jobs.',
    url_prefix='/job_api'
)


def query(sql, params=()):
    with current_app.app_context():
        mysql_config = {
            'user': current_app.config.get('MYSQL_DATABASE_USER'),
            'database': current_app.config.get('MYSQL_DATABASE_DB'),
            'password': current_app.config.get('MYSQL_DATABASE_PASSWORD'),
            'host': current_app.config.get('MYSQL_DATABASE_HOST'),
        }
        # table = current_app.config.get('MYSQL_DATABASE_TABLE'),

        cnx = mysql.connector.connect(**mysql_config)
        response = []
        cursor = cnx.cursor()
        try:
            cursor.execute(sql, params)
            if 'UPDATE' in sql or 'INSERT' in sql:
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


@jobs_api.route('/fetch_new', endpoint='jobs_fetch_new')
class NewJobList(Resource):
    """Fetch new Jobs for execution
    NOTE: The API accepts any run-uid but existing JobFlowId.
    So potentially we can fetch jobs for multiple clusters
    """

    @jobs_api.doc(params={
        'jobflowid': 'The canonical Run identifier for the EMR Cluster',
        'run_uid': 'The canonical Run identifier for the request',
        'limit': 'Max limit'
    })
    def get(self):
        try:
            request.get_json(force=True)
        except BadRequest:
            pass

        parser = reqparse.RequestParser()
        parser.add_argument('run_uid', required=False, type=str, default='',
                            help="The canonical identifier for the request")
        parser.add_argument('jobFlowId', required=False, type=str,
                            help="EMR jobFlowId")
        parser.add_argument('jobflowid', required=False, type=str,
                            help="EMR jobFlowId")
        parser.add_argument('limit', required=False, type=int, default=1,
                            help="Limit results")
        args = parser.parse_args()
        limit = max(int(args['limit']), 1)
        status_code = 500
        ret = []
        if isinstance(args.get('jobflowid', ''), str) and args.get('jobflowid', ''):
            args['jobFlowId'] = args.get('jobflowid')

        try:

            for row in query("SELECT * from jobs WHERE status in ('pending', 'PENDING') LIMIT 10"):
                ret.append({
                    **row,
                    'job_id': str(row['id']),
                    'job_uid': str(row['id']),
                    'run_uid': '1',
                    'job_status': row['status'],
                    'extra_run_id': '1',
                    'jobFlowId': args['jobFlowId'],
                    'created_at': row['created_at'].isoformat(),
                    'ttr': int(row['ttr'])
                })
            time.sleep(randint(1, 3))
            for elem in ret:
                query(
                    f"UPDATE jobs SET status='propogated' WHERE id={elem['job_uid']} AND status IN ( 'pending','PENDING')")

                query(f"INSERT INTO jobs (ttr, cmd) VALUES({randint(1, 300)},'sleep {randint(1, 390)}');")

            logger.info(f"New {len(ret)} jobs")
            status_code = 200
        except Exception as e:
            ret = {}
            error_type = type(e).__name__
            logger.debug(f"Got error {error_type} {e} {traceback.print_stack(limit=10)}")
            ret['error_msg'] = f"{error_type}: {e}"
            ret['has_error'] = True
        return ret, status_code

    post = get


@jobs_api.route('/fetch_cancel', endpoint='jobs_cancel')
class CancelJob(Resource):
    """Fetch cancelled Jobs
    """

    @jobs_api.doc(params={
        'jobflowid': 'The canonical Run identifier for the EMR Cluster',
        'limit': 'Max limit'
    })
    def get(self):
        try:
            request.get_json(force=True)
        except BadRequest:
            pass

        parser = reqparse.RequestParser()

        parser.add_argument('jobFlowId', required=False, type=str,
                            help="EMR jobFlowId")
        parser.add_argument('jobflowid', required=False, type=str,
                            help="EMR jobFlowId")
        parser.add_argument('limit', required=False, type=int, default=1,
                            help="Limit results")
        args = parser.parse_args()
        limit = max(int(args['limit']), 1)
        status_code = 500
        ret = []
        if isinstance(args.get('jobflowid', ''), str) and args.get('jobflowid', ''):
            args['jobFlowId'] = args.get('jobflowid')

        try:
            status_code = 200
            for row in query(f"SELECT * from jobs WHERE id > 99 AND status in ('running', 'RUNNING') LIMIT {limit}"):
                ret.append({
                    **row,
                    'job_id': row['id'],
                    'jobFlowId': args['jobFlowId'],
                    'created_at': row['created_at'].isoformat()
                })
            # if ret:
            time.sleep(randint(1, 3))
            logger.info(f"Cancel {len(ret)} jobs")

        except Exception as e:
            ret = {}
            error_type = type(e).__name__
            logger.debug(f"Got error {error_type} {e} {traceback.print_stack(limit=10)}")
            ret['error_msg'] = f"{error_type}: {e}"
            ret['has_error'] = True
        return ret, status_code

    post = get


@jobs_api.route('/runs', endpoint='jobs_run')
class RunJob(Resource):
    """Working with progress
    """

    @jobs_api.doc(params={
        'job_uid': 'The canonical identifier for the request',
        'run_uid': 'The canonical Run identifier for the request',
        'job_status': 'Job status',
        'previous_job_status': 'Previous Job status',
        'extra_run_id': "The extra identifier for the run"
    })
    def put(self):
        try:
            request.get_json(force=True)
        except BadRequest:
            pass
        logger.info("starting runs")

        parser = reqparse.RequestParser()
        parser.add_argument('job_uid', required=False, type=str, default='',
                            help="The canonical identifier for the request")
        parser.add_argument('extra_run_id', required=False, type=str, default='',
                            help="The canonical extra_run_id for the request")
        parser.add_argument('job_status', required=False, type=str, help="New Job Status")
        parser.add_argument('previous_job_status', required=False, type=str,
                            default='PENDING',
                            help="Old Job Status")
        parser.add_argument('run_uid', required=False, type=str, default='',
                            help="The Run canonical identifier for the request")

        args = parser.parse_args()
        job_uid = args['job_uid']
        extra_run_id = args['extra_run_id']
        run_uid = args['run_uid']
        job_status = args['job_status']
        previous_job_status = args['previous_job_status']
        status_code = 500
        ret = {
            "job_id": str(job_uid),
            "jobStatus": job_status,
            "extra_run_id": extra_run_id,
            "has_error": False,
            "error_msg": "",
            "run_uid": run_uid,
            "previous_job_status": f"{previous_job_status}",
            "job_name": "",
            "cmd": "",
            "parameters": []
        }
        try:
            status_code = 200
            query(f"UPDATE jobs SET status='{job_status}' WHERE id={args['job_uid']} AND status IN ( '{previous_job_status}','PENDING', 'propogated')" )

            for row in query(f"SELECT * from jobs WHERE id='{args['job_uid']}'"):
                _ret = {
                    **ret,
                    **row,
                    'job_id': row['id'],
                    'jobFlowId': args['jobFlowId'],
                    'created_at': row['created_at'].isoformat()
                }
            logger.info(f" update job {job_uid} with {_ret}")
        except Exception as e:
            ret = {}
            error_type = type(e).__name__
            logger.info(f"Got error {error_type} {e}")
            # logger.info(f"Got error {error_type} {e} {traceback.print_stack(limit=10)}")
            # ret['error_msg'] = f"{error_type}: {e}"
            ret['has_error'] = True
        return ret, status_code
