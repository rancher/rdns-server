import json
import requests

from common import BASE_URL


def test_server_apis():  # NOQA
    # test create
    url = build_url(BASE_URL, "", "")
    response = create_domain_test(url,
                                  {
                                      'fqdn': '',
                                      'hosts': ["1.1.1.1", "3.3.3.3"]
                                  })
    assert response != ""
    result = response.json()
    assert result['status'] == 200
    assert result['token'] != ""
    assert result['data']['fqdn'] != ""
    token = result['token']
    fqdn = result['data']['fqdn']
    expiration_time = result['data']['expiration']
    for host in result['data']['hosts']:
        assert host in ["1.1.1.1", "3.3.3.3"]

    # test create acme text record
    acme_url = build_url(BASE_URL, "/_acme-challenge." + fqdn, "/txt")
    response = create_domain_text_test(acme_url,
                                       token,
                                       {"text": "acme challenge record"})
    assert response != ""
    result = response.json()
    assert result['status'] == 200
    assert result['data']['fqdn'] != ""
    acme_expiration_time = result['data']['expiration']
    assert result['data']['text'] == "acme challenge record"

    # test update
    url = build_url(BASE_URL, "/" + fqdn, "")
    response = update_domain_test(url,
                                  token,
                                  {'fqdn': '', 'hosts': ["2.2.2.2"]})
    assert response != ""
    result = response.json()
    assert result['status'] == 200
    for host in result['data']['hosts']:
        assert host in ["2.2.2.2"]

    # test update acme text record
    acme_url = build_url(BASE_URL, "/_acme-challenge." + fqdn, "/txt")
    response = update_domain_test(acme_url,
                                  token,
                                  {"text": "acme challenge record updated"})
    assert response != ""
    result = response.json()
    assert result['status'] == 200
    assert result['data']['text'] == "acme challenge record updated"

    # test renew
    url = build_url(BASE_URL, "/" + fqdn, "/renew")
    response = renew_domain_test(url, token)
    assert response != ""
    result = response.json()
    assert result['status'] == 200
    assert result['data']['expiration'] > expiration_time

    # test renew acme text record
    acme_url = build_url(BASE_URL, "/_acme-challenge." + fqdn, "/txt")
    response = get_domain_test(acme_url, token)
    assert response != ""
    result = response.json()
    assert result['status'] == 200
    assert result['data']['expiration'] > acme_expiration_time

    # test delete
    url = build_url(BASE_URL, "/" + fqdn, "")
    response = delete_domain_test(url, token)
    assert response != ""
    result = response.json()
    assert result['status'] == 200

    # test delete acme text record
    acme_url = build_url(BASE_URL,
                         "/_acme-challenge." + fqdn, "/txt")
    response = delete_domain_test(acme_url, token)
    assert response != ""
    result = response.json()
    assert result['status'] == 200

    # check delete result
    url = build_url(BASE_URL, "/" + fqdn, "")
    response = get_domain_test(url, token)
    assert response != ""
    result = response.json()
    assert result['status'] == 200
    assert result['data'] == {}

    # check acme text record
    acme_url = build_url(BASE_URL, "/_acme-challenge." + fqdn, "/txt")
    response = get_domain_test(acme_url, token)
    assert response != ""
    result = response.json()
    assert result['status'] == 200
    assert result['data'] == {}


# This method creates the domain
def create_domain_test(url, data):
    headers = build_header("")
    response = requests.post(url, data=json.dumps(data), headers=headers)
    return response


# This method creates the domain text
def create_domain_text_test(url, token, data):
    headers = build_header(token)
    response = requests.post(url, data=json.dumps(data), headers=headers)
    return response


# This method gets the domain
def get_domain_test(url, token):
    headers = build_header(token)
    response = requests.get(url, params=None, headers=headers)
    return response


# This method deletes the domain
def delete_domain_test(url, token):
    headers = build_header(token)
    response = requests.delete(url, headers=headers)
    return response


# This method renews the domain
def renew_domain_test(url, token):
    headers = build_header(token)
    response = requests.put(url, data=None, headers=headers)
    return response


# This method updates the domain
def update_domain_test(url, token, data):
    headers = build_header(token)
    response = requests.put(url, data=json.dumps(data), headers=headers)
    return response


# build_url return request url
def build_url(base, fqdn, path):
    return '%s/domain%s%s' % (base, fqdn, path)


# build_header return request header
def build_header(token):
    if token == "":
        return {"Content-Type": "application/json",
                "Accept": "application/json"}

    return {"Content-Type": "application/json",
            "Accept": "application/json",
            "Authorization": 'Bearer %s' % token}