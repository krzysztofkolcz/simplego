Udało mi się zrobić ingestion.
Wysyłam teraz request retrieval:
POST na https://canary.als.services.cloud.sap/retrieval/v1/query
```
{
    "timeFrom": "2026-03-05T10:00:00.000Z",
    "timeTo": "2026-03-05T18:00:00.000Z",
    "source": "/test-region-kk/sap.kmscmk/776c68d1-2590-405f-924e-96f0cd284079",
    "xsapcustompartition": {},
    "dataFilter": {
      "tenantId": "dc904efa-7b0b-4487-b00e-ec37b6c69e28"
    },
    "eventType": "unauthenticatedRequest"
}
```
Dostaje response:
```
{
  "id": "09bcf977-bb03-4f3a-8a28-5830aa189efe"
}
```
Pobieram status - jest ok:
GET https://canary.als.services.cloud.sap/retrieval/v1/query/09bcf977-bb03-4f3a-8a28-5830aa189efe/status
```
HTTP/2 200 OK
{
  "id": "09bcf977-bb03-4f3a-8a28-5830aa189efe",
  "timeFrom": "2026-03-05T10:00:00",
  "timeTo": "2026-03-05T18:00:00",
  "filter": "{\"timeFrom\":\"2026-03-05T10:00:00\",\"timeTo\":\"2026-03-05T18:00:00\",\"source\":\"/test-region-kk/sap.kmscmk/776c68d1-2590-405f-924e-96f0cd284079\",\"eventType\":\"unauthenticatedRequest\",\"xsapcustompartition\":{},\"dataFilter\":{\"tenantId\":\"dc904efa-7b0b-4487-b00e-ec37b6c69e28\"}}",
  "createdAt": "2026-03-05T09:26:00.881+00:00",
  "status": "finished",
  "error": ""
}
```

Pobieram results - jest ok, ale empty:
GET https://canary.als.services.cloud.sap/retrieval/v1/query/09bcf977-bb03-4f3a-8a28-5830aa189efe/results
```
[
]
```
