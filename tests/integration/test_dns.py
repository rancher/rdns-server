import dns.resolver
import os

from test_rdns import build_url, \
    create_domain_test, \
    create_domain_text_test, \
    delete_domain_test

envs = os.environ
BASE_URL = envs.get('ENV_RDNS_ENDPOINT')


def test_dns():  # NOQA
    url = build_url(BASE_URL, "", "")
    response = create_domain_test(url,
                                  {
                                      'fqdn': '',
                                      'hosts': ["4.4.4.4"],
                                      'subdomain': {
                                          'sub1': ["5.5.5.5"],
                                      },
                                  })
    result = response.json()
    token = result['token']
    fqdn = result['data']['fqdn']

    # test query A record
    dns_query = dns.message.make_query(fqdn, 'A')
    response = dns.query.udp(dns_query, '8.8.8.8')
    for i in response.answer:
        for j in i.items:
            assert str(j) in ["4.4.4.4"]
    sub_query = dns.message.make_query('sub1.' + fqdn, 'A')
    response = dns.query.udp(sub_query, '8.8.8.8')
    for i in response.answer:
        for j in i.items:
            assert str(j) in ["5.5.5.5"]

    # test query TXT record
    acme_url = build_url(BASE_URL,
                         "/_acme-challenge." + fqdn, "/txt")
    response = create_domain_text_test(acme_url,
                                       token,
                                       {
                                           "text": "acme another record"
                                       })
    result = response.json()
    acme_fqdn = result['data']['fqdn']

    dns_query = dns.message.make_query(acme_fqdn, 'TXT')
    response = dns.query.udp(dns_query, '8.8.8.8')
    for i in response.answer:
        for j in i.items:
            acme_text = j.to_text()
            assert acme_text == '"acme another record"'

    url = build_url(BASE_URL, "/_acme-challenge." + fqdn, "/txt")
    delete_domain_test(url, token)

    url = build_url(BASE_URL, "/" + fqdn, "")
    delete_domain_test(url, token)
