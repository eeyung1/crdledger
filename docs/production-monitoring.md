# Production monitoring runbook

CRDLedger production runs on Render and uses Paystack for seller subscriptions.

## Health
- Public health endpoint: `/healthz`.
- A healthy response is HTTP 200 with `ok`.
- Render deploys should reach `live` before a release is treated as deployed.

## Payment signals
Search structured application logs for:
- `subscription activated`
- `webhook subscription activated`
- `subscription checkout initialization failed`
- `subscription payment verification failed`
- `subscription payment verification unavailable`
- `subscription activation failed`
- `webhook payment verification rejected`
- `webhook payment verification unavailable`
- `webhook subscription activation failed`

Successful callback/webhook processing is idempotent. Repeated delivery must not extend a subscription twice.

## What to investigate
1. Any sustained HTTP 5xx responses.
2. A failed deploy or a deploy that never becomes live.
3. Repeated Paystack verification/activation errors.
4. Unexpected growth in CPU or memory.
5. A completed Paystack charge whose CRDLedger payment record is not completed.

## Payment incident check
Use the Paystack reference shown in the user's Profile or Admin > Subscriptions. Compare that reference with application logs and Paystack before changing account state. Do not manually extend a subscription merely because a browser callback failed; the webhook may already have activated it.

## Release check
After a payment-related deploy:
1. Confirm deploy status is live.
2. Confirm `/healthz` returns 200.
3. Check recent logs for startup/database errors.
4. Check recent request/error trends.
5. Only then run the relevant production acceptance flow.
