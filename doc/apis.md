# API References

> CNAME feature only supported by `route53`

| API | Method | Header | Payload | Description |
| --- | ------ | ------ | ------- | ----------- |
| /v1/domain | POST | **Content-Type:** application/json <br/><br/> **Accept:** application/json | {"hosts": ["4.4.4.4", "2.2.2.2"], "subdomain": {"sub1": ["9.9.9.9","4.4.4.4"], "sub2": ["5.5.5.5","6.6.6.6"]}} | Create A Records |
| /v1/domain/&lt;FQDN&gt; | GET | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | - | Get A Records |
| /v1/domain/&lt;FQDN&gt; | PUT | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | {"hosts": ["4.4.4.4", "3.3.3.3"], "subdomain": {"sub1": ["9.9.9.9","4.4.4.4"], "sub3": ["5.5.5.5","6.6.6.6"]}} | Update A Records |
| /v1/domain/&lt;FQDN&gt; | DELETE | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | - | Delete A Records |
| /v1/domain/&lt;FQDN&gt;/txt | POST | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | {"text": "xxxxxx"} | Create TXT Record |
| /v1/domain/&lt;FQDN&gt;/txt | GET | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | - | Get TXT Record |
| /v1/domain/&lt;FQDN&gt;/txt | PUT | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | {"text": "xxxxxxxxx"} | Update TXT Record |
| /v1/domain/&lt;FQDN&gt;/txt | DELETE | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | - | Delete TXT Record |
| /v1/domain/&lt;FQDN&gt;/cname | POST | **Content-Type:** application/json <br/><br/> **Accept:** application/json | {"cname": "xxxxxx"} | Create CNAME Record |
| /v1/domain/&lt;FQDN&gt;/cname | GET | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | - | Get CNAME Record |
| /v1/domain/&lt;FQDN&gt;/cname | PUT | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | {"cname": "xxxxxxxxx"} | Update CNAME Record |
| /v1/domain/&lt;FQDN&gt;/cname | DELETE | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | - | Delete CNAME Record |
| /v1/domain/&lt;FQDN&gt;/renew | PUT | **Content-Type:** application/json <br/><br/> **Accept:** application/json <br/><br/> **Authorization:** Bearer &lt;Token&gt; | - | Renew Records |
| /metrics | GET | - | - | Prometheus metrics |
