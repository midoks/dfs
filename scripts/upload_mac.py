#!/usr/bin/env python
# encoding: utf-8
"""
上传DEMO
"""

'''
ffmpeg
'''

import hashlib
import os
import time
import datetime
import traceback
import sys
import json
import socket
import threading
import subprocess
import mimetypes


reload(sys)
sys.setdefaultencoding("utf8")

sys.path.append('/usr/local/lib/python2.7/site-packages')
import requests


runDir = os.getcwd()

ffmpeg_cmd = "/usr/local/bin/ffmpeg"


def execShell(cmdstring, cwd=None, timeout=None, shell=True):

    if shell:
        cmdstring_list = cmdstring
    else:
        cmdstring_list = shlex.split(cmdstring)
    if timeout:
        end_time = datetime.datetime.now() + datetime.timedelta(seconds=timeout)

    sub = subprocess.Popen(cmdstring_list, cwd=cwd, stdin=subprocess.PIPE,
                           shell=shell, bufsize=4096, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

    while sub.poll() is None:
        time.sleep(0.1)
        if timeout:
            if end_time <= datetime.datetime.now():
                raise Exception("Timeout：%s" % cmdstring)

    return sub.communicate()


def fg_transfer_ts_cmd(file, to_file):
    cmd = ffmpeg_cmd + ' -y -i ' + file + \
        ' -s 480x360 -vcodec copy -acodec copy -vbsf h264_mp4toannexb ' + to_file
    return cmd


def fg_m3u8_cmd(ts_file, m3u8_file, to_file):
    cmd = ffmpeg_cmd + ' -y -i ' + ts_file + ' -c copy -map 0 -f segment -segment_list ' + \
        m3u8_file + ' -segment_time 10 ' + to_file
    return cmd

# print
# cmd =fg_transfer_ts_cmd(runDir+"/video/a.mp4",runDir+"/tmp/a.ts")

# cmd = fg_m3u8_cmd(runDir + "/video/a.mp4", runDir +
#                   "/tmp/mm/a.m3u8", runDir + "/tmp/mm/%010d.ts")
# print cmd
# print os.system(cmd)


def is_m3u8_file(f):
    a = f.split(".")
    if (a[1] == 'm3u8'):
        return True
    return False


sourePath = runDir + "/video/a.mp4"


cmd = "md5 " + sourePath + " | cut -d ' ' -f4"
print cmd
data = execShell(cmd)
md5file = data[0].strip()

url = "http://127.0.0.1:8081/upload"
headers = {'User-Agent': 'Chrome/71.0.3578.98 Safari/537.36'}
for f in os.listdir(runDir + "/tmp/mm/"):
    if f[0:1] == '.':
        continue
    # if is_m3u8_file(f):
    #     continue
    print f
    fullPath = runDir + "/tmp/mm/" + f
    # print mimetypes.guess_type(fullPath)[0]
    files = {
        # image或者file
        'file': (f, open(fullPath, 'rb'), mimetypes.guess_type(fullPath)[0])
    }
    r = requests.post(
        url=url, data={"scene": "m3u8", "group_md5": md5file, "fixed_dir": "sss"}, headers=headers, files=files)
    r.raise_for_status()
    print(r.text)
    print(r.status_code)
