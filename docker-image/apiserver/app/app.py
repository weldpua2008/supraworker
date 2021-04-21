# -*- coding: utf-8 -*-
"""
Main file
"""
import sys
import os
import logging
from flask import Flask
from flask_restx.apidoc import apidoc
from app.resources.logs import logs_page
from app.resources.jobs import job_page
logging.info("Starting Flask...")
logging.info(f"Load FLASK config {os.getenv('APP_SETTINGS', 'flaskconfig.ProductionConfig')}")

app = Flask(__name__)
app.config.from_object(os.getenv('APP_SETTINGS', 'flaskconfig.ProductionConfig'))

logging.info("Register Blueprint")


apidoc.static_url_path = "{}/swagger/ui".format(app.config['URL_PREFIX'])

app.register_blueprint(job_page, url_prefix="{}/jobs".format(app.config['URL_PREFIX']))
app.register_blueprint(logs_page, url_prefix="{}/logs".format(app.config['URL_PREFIX']))

logging.info("FINISHED INITIALIZATION")
