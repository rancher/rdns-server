import dns.resolver

from common import BASE_URL, IP, NAMESERVER_PORT
from test_basic import build_url, \
    create_domain_test, \
    create_domain_text_test, \
    delete_domain_test


def test_core_dns():  # NOQA
    url = build_url(BASE_URL, "", "")
    response = create_domain_test(url, {'fqdn': '', 'hosts': ["4.4.4.4"]})
    result = response.json()
    token = result['token']
    fqdn = result['data']['fqdn']

    # test query A record
    dns_query = dns.message.make_query(fqdn, 'A')
    response = dns.query.udp(dns_query, IP, port=NAMESERVER_PORT)
    for i in response.answer:
        for j in i.items:
            assert str(j) in ["4.4.4.4"]

    acme_url = build_url(BASE_URL,
                         "/_acme-challenge." + fqdn, "/txt")
    response = create_domain_text_test(acme_url,
                                       token,
                                       {
                                           "text": "acme challenge another record"
                                       })
    result = response.json()
    acme_fqdn = result['data']['fqdn']

    # test query TXT record
    dns_query = dns.message.make_query(acme_fqdn, 'TXT')
    response = dns.query.udp(dns_query, IP, port=NAMESERVER_PORT)
    for i in response.answer:
        for j in i.items:
            acme_text = j.to_text()
            assert acme_text == '"acme challenge another record"'

    url = build_url(BASE_URL, "/_acme-challenge." + fqdn, "/txt")
    delete_domain_test(url, token)

    url = build_url(BASE_URL, "/" + fqdn, "")
    delete_domain_test(url, token)