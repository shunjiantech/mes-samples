import base64
import hashlib
import hmac
from datetime import datetime
from urllib.parse import urlparse

import requests


class Sdk:
    def __init__(self, base_url, app_key, app_secret):
        self.session = requests.session()
        self.base_url = base_url
        self.app_key = app_key
        self.app_secret = app_secret

    # 获取被试品信息
    def get_devices(self, testStationId: int, qrcode: str):
        url = "{0}/devices?test_station_id={1}&qrcode={2}".format(self.base_url, testStationId, qrcode)

        req = requests.Request(method='GET', url=url)

        prep = self.session.prepare_request(req)

        self.sign_request(prep)

        resp = self.session.send(prep)

        if resp.status_code != 200:
            print("[Err] status code: {0}".format(resp.status_code))

        print(resp.text)

    # 保存试验数据
    def save_test_data(self, qrcode: str, json_obj: object):
        url = "{0}/devices/{1}/test_data".format(self.base_url, qrcode)

        req = requests.Request(method='POST', url=url, json=json_obj)

        prep = self.session.prepare_request(req)

        self.sign_request(prep)

        resp = self.session.send(prep)

        if resp.status_code != 200:
            print("[Err] status code: {0}".format(resp.status_code))

        print(resp.text)

    # 上传试验图片
    def upload_image(self, image_path: str):
        url = "{0}/uploads".format(self.base_url)

        files = {
            'file': open(image_path, 'rb')
        }

        req = requests.Request(method='POST', url=url, files=files)

        prep = self.session.prepare_request(req)

        self.sign_request(prep)

        resp = self.session.send(prep)

        if resp.status_code != 200:
            print("[Err] status code: {0}".format(resp.status_code))

        print(resp.text)

    # 签名
    def sign_request(self, prep: requests.PreparedRequest):
        uri = urlparse(prep.url)

        query = ''
        if uri.query:
            query_dict = {}
            for q in uri.query.split('&'):
                qs = q.split('=')
                query_dict[qs[0]] = qs[1]

            query_dict = sorted(query_dict.items(), key=lambda item: item[0])
            query = '&'.join([f'{k}={v}' for k, v in query_dict])

        content_type = ''
        if 'content-type' in prep.headers:
            content_type = prep.headers['content-type']

        content_md5 = ''
        if prep.method == "POST" or prep.method == "PUT":
            if prep.body:
                content_md5 = self.md5(prep.body)

        timestamp = str(int(datetime.now().timestamp() * 1e3))

        string_to_sign = \
            prep.method + '\n' + \
            uri.path + '\n' + \
            query + '\n' + \
            content_type + '\n' + \
            content_md5 + '\n' + \
            timestamp + '\n' + \
            self.app_secret

        signature = self.hmac_sha1(self.app_key.encode('utf-8'), string_to_sign.encode("utf-8"))

        prep.headers['X-AppKey'] = self.app_key
        prep.headers['X-Timestamp'] = timestamp
        prep.headers['X-Signature'] = signature

    @staticmethod
    def md5(data: bytes):
        h = hashlib.md5()
        h.update(data)
        return base64.b64encode(h.digest()).decode("utf-8")

    @staticmethod
    def hmac_sha1(key: bytes, data: bytes):
        h = hmac.new(key, data, "SHA1")
        return base64.b64encode(h.digest()).decode("utf-8")


sdk = Sdk('http://127.0.0.1:8080/api/open', 'appKey', 'appSecret')
sdk.get_devices(67, '20210901103050484')
sdk.save_test_data('20210901103050484', [
    {
        "test_item_id": 10000,
        "test_data": [
            {
                "item1": 1,
                "item2": "2",
                "item3": True
            }
        ]
    }
])
sdk.upload_image('./baidu.png')
