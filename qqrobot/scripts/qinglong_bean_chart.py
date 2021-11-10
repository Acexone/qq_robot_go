import datetime
import json
import logging
import re
import time
from datetime import timedelta, timezone
from urllib.parse import urlencode

import requests
import requests.adapters
from PIL import Image, ImageFont, ImageDraw
from prettytable import PrettyTable

FONT_FILE = f'font/jet.ttf'
BEAN_IMG = f'bean.jpeg'
CHART_IMG = f'chart.jpeg'
AUTH_JSON = "D:/_codes/js/qinglong/data/config/auth.json"
QL_API_ADDR = "http://127.0.0.1:5700/api"
# QUICK_CHART_ADDR = "https://quickchart.io"
QUICK_CHART_ADDR = "http://127.0.0.1:5703"

logger = logging.getLogger(__name__)

SHA_TZ = timezone(
    timedelta(hours=8),
    name='Asia/Shanghai',
)
requests.adapters.DEFAULT_RETRIES = 5
session = requests.session()
session.keep_alive = False


def get_bean(account_idx: int):
    res = get_bean_data(account_idx)
    if res['code'] == 200:
        creat_bean_count(res['data'][3], res['data'][0], res['data'][1], res['data'][2][1:])
        print(f'您的账号 {account_idx} 收支情况 统计表格 已保存到 {BEAN_IMG}')


def get_chart(account_idx: int):
    res = get_bean_data(int(account_idx))
    if res['code'] == 200:
        creat_chart(res['data'][3], f'账号{str(account_idx)}', res['data'][0], res['data'][1], res['data'][2][1:])
        print(f'您的账号 {account_idx} 收支情况 统计图 已保存到 {CHART_IMG}')


def get_bean_data(i):
    try:
        cookies = get_cks(AUTH_JSON)
        if cookies:
            ck = cookies[i - 1]
            beans_res = get_beans_7days(ck)
            beantotal = get_total_beans(ck)
            if beans_res['code'] != 200:
                return beans_res
            else:
                beans_in, beans_out = [], []
                beanstotal = [int(beantotal), ]
                for i in beans_res['data'][0]:
                    beantotal = int(
                        beantotal) - int(beans_res['data'][0][i]) - int(beans_res['data'][1][i])
                    beans_in.append(int(beans_res['data'][0][i]))
                    beans_out.append(int(str(beans_res['data'][1][i]).replace('-', '')))
                    beanstotal.append(beantotal)
            return {'code': 200, 'data': [beans_in[::-1], beans_out[::-1], beanstotal[::-1], beans_res['data'][2][::-1]]}
    except Exception as e:
        logger.error(str(e))


def get_cks(ckfile):
    ck_reg = re.compile(r'pt_key=\S*?;.*?pt_pin=\S*?;')

    with open(ckfile, 'r', encoding='utf-8') as f:
        auth = json.load(f)
    lines = str(env_manage_QL('search', 'JD_COOKIE', auth['token']))

    cookies = ck_reg.findall(lines)
    for ck in cookies:
        if ck == 'pt_key=xxxxxxxxxx;pt_pin=xxxx;':
            cookies.remove(ck)
            break
    return cookies


def env_manage_QL(fun, envdata, token):
    url = f'{QL_API_ADDR}/envs'
    headers = {
        'Authorization': f'Bearer {token}'
    }
    try:
        if fun == 'search':
            params = {
                't': int(round(time.time() * 1000)),
                'searchValue': envdata
            }
            res = requests.get(url, params=params, headers=headers).json()
        elif fun == 'add':
            data = {
                'name': envdata['name'],
                'value': envdata['value'],
                'remarks': envdata['remarks'] if 'remarks' in envdata.keys() else ''
            }
            res = requests.post(url, json=[data], headers=headers).json()
        elif fun == 'edit':
            data = {
                'name': envdata['name'],
                'value': envdata['value'],
                '_id': envdata['_id'],
                'remarks': envdata['remarks'] if 'remarks' in envdata.keys() else ''
            }
            res = requests.put(url, json=data, headers=headers).json()
        elif fun == 'disable':
            data = [envdata['_id']]
            res = requests.put(url + '/disable', json=data,
                               headers=headers).json()
        elif fun == 'enable':
            data = [envdata['_id']]
            res = requests.put(url + '/enable', json=data,
                               headers=headers).json()
        elif fun == 'del':
            data = [envdata['_id']]
            res = requests.delete(url, json=data, headers=headers).json()
        else:
            res = {'code': 400, 'data': '未知功能'}
    except Exception as e:
        res = {'code': 400, 'data': str(e)}
    finally:
        return res


def get_beans_7days(ck):
    try:
        day_7 = True
        page = 0
        headers = {
            "Host": "api.m.jd.com",
            "Connection": "keep-alive",
            "charset": "utf-8",
            "User-Agent": "Mozilla/5.0 (Linux; Android 10; MI 9 Build/QKQ1.190825.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/78.0.3904.62 XWEB/2797 MMWEBSDK/201201 Mobile Safari/537.36 MMWEBID/7986 MicroMessenger/8.0.1840(0x2800003B) Process/appbrand4 WeChat/arm64 Weixin NetType/4G Language/zh_CN ABI/arm64 MiniProgramEnv/android",
            "Content-Type": "application/x-www-form-urlencoded;",
            "Accept-Encoding": "gzip, compress, deflate, br",
            "Cookie": ck,
            "Referer": "https://servicewechat.com/wxa5bf5ee667d91626/141/page-frame.html",
        }
        days = []
        for i in range(0, 7):
            days.append(
                (datetime.date.today() - datetime.timedelta(days=i)).strftime("%Y-%m-%d"))
        beans_in = {key: 0 for key in days}
        beans_out = {key: 0 for key in days}
        while day_7:
            page = page + 1
            resp = session.get("https://api.m.jd.com/api", params=gen_params(page),
                               headers=headers, timeout=100).text
            res = json.loads(resp)
            if res['resultCode'] == 0:
                for i in res['data']['list']:
                    for date in days:
                        if str(date) in i['createDate'] and i['amount'] > 0:
                            beans_in[str(date)] = beans_in[str(
                                date)] + i['amount']
                            break
                        elif str(date) in i['createDate'] and i['amount'] < 0:
                            beans_out[str(date)] = beans_out[str(
                                date)] + i['amount']
                            break
                    if i['createDate'].split(' ')[0] not in str(days):
                        day_7 = False
            else:
                return {'code': 400, 'data': res}
        return {'code': 200, 'data': [beans_in, beans_out, days]}
    except Exception as e:
        logger.error(str(e))
        return {'code': 400, 'data': str(e)}


def gen_params(page):
    body = gen_body(page)
    params = {
        "functionId": "jposTradeQuery",
        "appid": "swat_miniprogram",
        "client": "tjj_m",
        "sdkName": "orderDetail",
        "sdkVersion": "1.0.0",
        "clientVersion": "3.1.3",
        "timestamp": int(round(time.time() * 1000)),
        "body": json.dumps(body)
    }
    return params


def gen_body(page):
    body = {
        "beginDate": datetime.datetime.utcnow().replace(tzinfo=timezone.utc).astimezone(SHA_TZ).strftime("%Y-%m-%d %H:%M:%S"),
        "endDate": datetime.datetime.utcnow().replace(tzinfo=timezone.utc).astimezone(SHA_TZ).strftime("%Y-%m-%d %H:%M:%S"),
        "pageNo": page,
        "pageSize": 20,
    }
    return body


def get_total_beans(ck):
    try:
        headers = {
            "Host": "wxapp.m.jd.com",
            "Connection": "keep-alive",
            "charset": "utf-8",
            "User-Agent": "Mozilla/5.0 (Linux; Android 10; MI 9 Build/QKQ1.190825.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/78.0.3904.62 XWEB/2797 MMWEBSDK/201201 Mobile Safari/537.36 MMWEBID/7986 MicroMessenger/8.0.1840(0x2800003B) Process/appbrand4 WeChat/arm64 Weixin NetType/4G Language/zh_CN ABI/arm64 MiniProgramEnv/android",
            "Content-Type": "application/x-www-form-urlencoded;",
            "Accept-Encoding": "gzip, compress, deflate, br",
            "Cookie": ck,
        }
        jurl = "https://wxapp.m.jd.com/kwxhome/myJd/home.json"
        resp = session.get(jurl, headers=headers, timeout=100).text
        res = json.loads(resp)
        return res['user']['jingBean']
    except Exception as e:
        logger.error(str(e))


def creat_bean_count(date, beansin, beansout, beanstotal):
    tb = PrettyTable()
    tb.add_column('DATE', date)
    tb.add_column('BEANSIN', beansin)
    tb.add_column('BEANSOUT', beansout)
    tb.add_column('TOTAL', beanstotal)
    font = ImageFont.truetype(FONT_FILE, 18)
    im = Image.new("RGB", (500, 260), (244, 244, 244))
    dr = ImageDraw.Draw(im)
    dr.text((10, 5), str(tb), font=font, fill="#000000")
    im.save(BEAN_IMG)


def creat_chart(xdata, title, bardata, bardata2, linedate):
    qc = QuickChart()
    qc.background_color = '#fff'
    qc.width = "1000"
    qc.height = "600"
    qc.config = {
        "type": "bar",
        "data": {
            "labels": xdata,
            "datasets": [
                {
                    "label": "IN",
                    "backgroundColor": [
                        "rgb(255, 99, 132)",
                        "rgb(255, 159, 64)",
                        "rgb(255, 205, 86)",
                        "rgb(75, 192, 192)",
                        "rgb(54, 162, 235)",
                        "rgb(153, 102, 255)",
                        "rgb(255, 99, 132)"
                    ],
                    "yAxisID": "y1",
                    "data": bardata
                },
                {
                    "label": "OUT",
                    "backgroundColor": [
                        "rgb(255, 99, 132)",
                        "rgb(255, 159, 64)",
                        "rgb(255, 205, 86)",
                        "rgb(75, 192, 192)",
                        "rgb(54, 162, 235)",
                        "rgb(153, 102, 255)",
                        "rgb(255, 99, 132)"
                    ],
                    "yAxisID": "y1",
                    "data": bardata2
                },
                {
                    "label": "TOTAL",
                    "type": "line",
                    "fill": False,
                    "backgroundColor": "rgb(201, 203, 207)",
                    "yAxisID": "y2",
                    "data": linedate
                }
            ]
        },
        "options": {
            "plugins": {
                "datalabels": {
                    "anchor": 'end',
                    "align": -100,
                    "color": '#666',
                    "font": {
                        "size": 20,
                    }
                },
            },
            "legend": {
                "labels": {
                    "fontSize": 20,
                    "fontStyle": 'bold',
                }
            },
            "title": {
                "display": True,
                "text": f'{title}   收支情况',
                "fontSize": 24,
            },
            "scales": {
                "xAxes": [{
                    "ticks": {
                        "fontSize": 24,
                    }
                }],
                "yAxes": [
                    {
                        "id": "y1",
                        "type": "linear",
                        "display": False,
                        "position": "left",
                        "ticks": {
                            "max": int(int(max([max(bardata), max(bardata2)]) + 100) * 2)
                        },
                        "scaleLabel": {
                            "fontSize": 20,
                            "fontStyle": 'bold',
                        }
                    },
                    {
                        "id": "y2",
                        "type": "linear",
                        "display": False,
                        "ticks": {
                            "min": int(min(linedate) * 2 - (max(linedate)) - 100),
                            "max": int(int(max(linedate)))
                        },
                        "position": "right"
                    }
                ]
            }
        }
    }
    qc.to_file(CHART_IMG)


class QuickChart:
    def __init__(self):
        self.config = None
        self.width = 500
        self.height = 300
        self.background_color = '#ffffff'
        self.device_pixel_ratio = 1.0
        self.format = 'png'
        self.version = '2.9.4'
        self.key = None
        # self.scheme = 'https'
        # self.host = 'quickchart.io'

    def is_valid(self):
        return self.config is not None

    def get_url_base(self):
        return QUICK_CHART_ADDR

    def get_url(self):
        if not self.is_valid():
            raise RuntimeError(
                'You must set the `config` attribute before generating a url')
        params = {
            'c': dump_json(self.config) if type(self.config) == dict else self.config,
            'w': self.width,
            'h': self.height,
            'bkg': self.background_color,
            'devicePixelRatio': self.device_pixel_ratio,
            'f': self.format,
            'v': self.version,
        }
        if self.key:
            params['key'] = self.key
        return '%s/chart?%s' % (self.get_url_base(), urlencode(params))

    def _post(self, url):
        try:
            import requests
        except:
            raise RuntimeError('Could not find `requests` dependency')

        postdata = {
            'chart': dump_json(self.config) if type(self.config) == dict else self.config,
            'width': self.width,
            'height': self.height,
            'backgroundColor': self.background_color,
            'devicePixelRatio': self.device_pixel_ratio,
            'format': self.format,
            'version': self.version,
        }
        if self.key:
            postdata['key'] = self.key
        st = time.time()
        resp = requests.post(url, json=postdata)
        if resp.status_code != 200:
            raise RuntimeError(
                'Invalid response code from chart creation endpoint')
        return resp

    def get_short_url(self):
        resp = self._post('%s/chart/create' % self.get_url_base())
        parsed = json.loads(resp.text)
        if not parsed['success']:
            raise RuntimeError(
                'Failure response status from chart creation endpoint')
        return parsed['url']

    def get_bytes(self):
        resp = self._post('%s/chart' % self.get_url_base())
        return resp.content

    def to_file(self, path):
        content = self.get_bytes()
        with open(path, 'wb') as f:
            f.write(content)


FUNCTION_DELIMITER_RE = re.compile('\"__BEGINFUNCTION__(.*?)__ENDFUNCTION__\"')


class QuickChartFunction:
    def __init__(self, script):
        self.script = script

    def __repr__(self):
        return self.script


def serialize(obj):
    if isinstance(obj, QuickChartFunction):
        return '__BEGINFUNCTION__' + obj.script + '__ENDFUNCTION__'
    if isinstance(obj, (datetime.date, datetime.datetime)):
        return obj.isoformat()
    return obj.__dict__


def dump_json(obj):
    ret = json.dumps(obj, default=serialize, separators=(',', ':'))
    ret = FUNCTION_DELIMITER_RE.sub(
        lambda match: json.loads('"' + match.group(1) + '"'), ret)
    return ret


def demo():
    get_bean(1)
    get_chart(1)


if __name__ == '__main__':
    # re: 增加解析参数，确定调用哪个函数，以及账号id
    demo()
    pass
