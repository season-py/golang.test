#coding=utf-8
#author:season
import re
import urllib
import urllib2
import cookielib
from collections import defaultdict
from BeautifulSoup import BeautifulSoup
'''
just for fun  
'''
cookie = urllib2.HTTPCookieProcessor(cookielib.LWPCookieJar())
opener = urllib2.build_opener(cookie)
urllib2.install_opener(opener)

session = {'SBVerifyCode': 'byJlzq7u9JpWkPAIxPgD/1P51sY=',
    '.SPBForms': 'B0BEEB669AC25A5CF293539262A5C7649BAB05DEF14A3FD5B4E13E090CFFFB73B941F9024140559C8A869C8CC0983A74148BD97E9FC41A48CB60972BDF54A6EC64C156A93178B3ED'
}

post = {'email': 'haishanzhang@cyou-inc.com',
    'body': 'splider man',
    'bringCount': '0'
    }

def setcookie(func):
    def wrapper(req):
        if not isinstance(req, urllib2.Request):
            req = urllib2.Request(req)
        req.add_header('Cookie', ';'.join(['='.join(i) for i in session.items()]))
        return func(req)
    return wrapper

@setcookie
def request(req):
    return urllib2.urlopen(req).read()

def go():
    events_pattern = re.compile('href="(\/whatEvents.aspx\/T-[0-9]+)"')
    members_pattern = re.compile(u'([0-9]+)人参加.*?([0-9]+)条留言')
    already_req_url = defaultdict(int)
    domain = 'http://10.5.17.74'
    url = domain + '/c/musle/whatEvents.aspx'

    for item in events_pattern.findall(request(url)):#+['/whatEvents.aspx/T-115']:
        if already_req_url[item] == 1:
            continue
        title, content = None, None
        try:
            already_req_url[item] += 1
            content = BeautifulSoup(request(domain+item))
            title = content.title.text
        except urllib2.HTTPError:
            continue
        status = content.find('div', {'class': 'spb-event-countdown'}).text
        if status == u'此活动已结束':
            continue
        messages = content.find('div', {'class': 'spb-event-count'}).text
        members = members_pattern.findall(messages)[0][0]
        if int(members) >= 32:
            print '活动超载，孩纸，洗洗睡吧！', '[%s]' % title
        else:
            print '有机可乘~', '[%s]' % title
            join_content = content.find('a', text=u'我要报名')
            if not join_content:
                print '榜上有名啦！'
                continue
            content = BeautifulSoup(
                request(domain+join_content.findParent().get('href', ''))
                )
            ipost = urllib.urlencode(post)
            req = urllib2.Request(
                domain+content.find('form').get('action'),
                data=ipost)
            content = BeautifulSoup(request(req))
            print content.find('div', {'class': 'tn-helper-flowfix'}).text



if __name__ == '__main__':
    go()