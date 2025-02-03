# Super SMS Bridge
A middleware for forwarding SMS to Telegram using Super Chat. Compatible with SmsForwarder Webhook.  

![demo](./docs/eng-demo.png)

## SmsForwarder Compatible

> [!WARNING]  
> SMS Reply is not available due to SmsForwarder's restrictions.

**Params**
```json
{
    "sender": "[from]",
    "text": "[org_content]",
    "timestamp": "[timestamp]",
    "sign": "[sign]"
}
```

![setup](./docs/eng-sms-forwarder-setup.png)