# -*- coding: utf-8 -*-
# import server
# import multiprocessing
# from prometheus_client import multiprocess

# workers = multiprocessing.cpu_count() * 2 + 1
# workers = 1
proc_name = 'supraworker-simple-api'
bind = ["0.0.0.0:8080", "0.0.0.0:8084"]
threads = 2

# def child_exit(server, worker):
#     multiprocess.mark_process_dead(worker.pid)
