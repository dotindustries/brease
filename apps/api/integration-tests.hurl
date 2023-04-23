GET http://localhost:4400/
HTTP 200

GET http://localhost:4400/contextIDTest/rules
Authorization: asdf
HTTP 200

POST http://localhost:4400/contextIDTest/rules/add
Authorization: asdf
```json
{
	"rule": {
		"action": "string",
		"description": "string",
		"expression": {
			"and": {
				"expression": [
					{
						"condition": {
							"key": "field_id",
							"type": "px",
							"stringValue": "some Prefix"
						}
					},
					{
						"condition": {
							"key": "field_id",
							"type": "sx",
							"stringValue": "endsBadly"
						}
					}
				]
			}
		},
		"id": "asdf",
		"target": {
			"target": "string",
			"type": "string",
			"value": "ZXhhbXBsZQ=="
		}
	}
}
```
HTTP 200