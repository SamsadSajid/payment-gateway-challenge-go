# 2. implement payment gateway api interface

Date: 2024-07-22

## Status

Accepted

## Context

We want to implement an payment gateway for the merchant to make payment.
This payment gateway will be responsible for taking a merchant request for
a payment and return a relevenat response to the user based on the outcome
of the action.

## Decision

### Assumptions made

#### Product level assumptions
1. The APi Gateway will support only REST. gRPC, graphQL is out of scope
2. For the given task, authentication and authorization service won't be implemented. We assume that all requests that lands in the payment gateway service went through authentication and authorization by the company's identity service.
3. For the given task, rate limiter and throttling mechanisms won't be implemented. We assume that the payment service can utilise the existing company framework for such
4. A merchant will be able to make a payment request via the API gateway. The API gateway will call the acquired bank to capture the payment.
5. The payment gateway will not be responsible for any business decisions on capturing payment. It should the responsibility of the acquired bank.
6. The payment gateway will not do any currency conversion
7. The `amount` field will always be two digit after the decimal in merchant system. Therefore, when they make a request to the payment gateway, the same amount can be converted to the exact amount by the bank. Example: 
   - `0.01` will be sent as `1`
   - `10.50` will be sent as `1050`
   - `10.667` will be sent as `1067`

#### Functional requirements assumptions
1. The payment gateway will return two types of response. Success and error response.
2. A success response is defined as when the merchant's request has been sent to the acquired bank and the bank returns a HTTP-200 status code. For a success response, the payment gateway will send the [PostPaymentResponse](/internal/models/payment.go) type
3. An error response is defined as when a user makes an invalid requests or the acquired bank does not return a HTTP 200 response i.e., bank server unavailable etc.
4. It was not clear from the requirements what does "Ensure your submission validates against no more than 3 currency codes" mean? So, I validate the the user request with three popular currency a) GBP b) USD and c) EUR
5. The `amount` field in the merchan request will fit within 32 bit of integer type.
6. We will use `string` type to define the card number. The trade-off between `string` and `int32` is that if a card is `19` digit long, and each digit can be between `0-9`, it will require `10^19` bits. In this case we will need `int64` type. However, using `string` type is easy to deduce the implementation details

#### Implementation details
1. The [PostPaymentResponse](/internal/models/payment.go) type will 
    - not contain the `Cvv` and `BankAuthorizationCode` to the merchant
    - return only the last 4 digits of the card number to comply with policies
    - return `Authorized` status if the bank returns `authorized: true`
    - return `Declined` status if the bank returns `authorized: false`
2. If the merchant makes an invalid request or if the acquired bank does not return HTTP 200 after retry has been exhausted, the merchant will receive HTTP 400 code
3. The payment gateway will send an attribute `app_code` in the error response. A merchant can reconcile with the customer service based on the `app_code`.
4. If the bank returns `http.StatusServiceUnavailable || http.StatusTooManyRequests`, the payment gateway will retry for `3` seconds with exponential backoff. After `3` sedonds the gateway will return an error response to the merchant with app code `ErrBankResponseStatusCodeNon200`.
5. If bank returns any other status code, the payment gateway will return `ErrRequestRejected` app_code.
6. The datastore(in-memory) will only persist the last 4 digits of the card number to comply with the policy.
7. It was not clear from the requirements what does "Ensure your submission validates against no more than 3 currency codes" mean? So, I validate the the user request with three popular currency a) GBP b) USD and c) EUR

#### Future features
1. Publish all events to Kafka for reconciliation and analytics purposes
2. Create an app_code table for customer_service/Care team

## Consequences

What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated.

With the proposed solution we can achieve implementing the payment gateway. It enables merchants to capture payment.

### Risk factors
1. Bank captures payment but times out returning a response to the gateway. In this behaviour, the payment gateway will return an error response to the user. The user can retry via the payment gateway. In this case the payment gateway will send the request again back to the back. The bank can reject this request if it correctly processed the previous request or create a new one. However, this may impact UX.
2. 