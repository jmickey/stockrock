# Stockrock - Stock Ticker Service

Stockrock is an API that retrieves stock data for the last N trading days and provides an average closing price.

## Note

Based on the task instructions to look up "a fixed number of closing prices of a specific stock" my interpretation was that the `NDAYS` value referred to the number of **trading days**, rather than the number of calendar days. Therefore my solution does not take into account weekends and holidays, rather it returns a set number of datapoints based on the provided `NDAYS` variable, along with the average closing price for the datapoints.

## Requirements

- Go v1.20.1+
- Docker
- Kubectl
- Alphavantage API key

## Build/Run

To build & run with docker

```
docker build -t <org>/stockrock:1.0.0 .
docker run --rm -e API_KEY=<your-api-key> -e NDAYS=7 -e SYMBOL=MSFT -e HOST=0.0.0.0 -p 8080:8080 <org>/stockrock:1.0.0
```

Kubernetes:

_Note: The provided kubernetes ingress requires the Nginx ingress controller to be deployed to the cluster._

Create the namespace:

```
kubectl apply -f ./k8s/1-namespace.yaml
```

Create a secret:

1. Convert your API to a base64 string
  ```
  echo <YOUR_API_KEY> | base64 --encode
  ```
2. Create `secret.yaml` in `./k8s/` and use the following template (make sure to paste in your base64 encoded API_KEY)
  ```
  apiVersion: v1
  kind: Secret
  metadata:
    name: stockrock-apikey
    namespace: platform
  type: Opaque
  data:
    apiKey: <YOUR_BASE64_API_KEY>
  ```
3. Apply the files to the cluster:
  ```
  kubectl apply -f ./k8s
  ```

If using `kind` and you can use the `cluster.yaml` config file to ensure that the port is correctly exposed via localhost.

## Endpoints

There are 2 endpoints available:

1. `/healthz` - health check endpoint. Returns 200 if the service is up.
2. `/api/stock-info` - Returns information about the specified ticker for N days, as configured via the services environment variables.

Example output (shortened):

```
{
  "last_refeshed": "Tue Feb 28 06:30:15 2023",
  "days": 7,
  "symbol": "MSFT",
  "average_closing_price": "254.08",
  "stock_time_series": [
    {
      "date": "2023-02-27T00:00:00-05:00",
      "open": "252.46",
      "high": "252.82",
      "low": "249.39",
      "close": "250.16",
      "volume": "21190042"
    },
    {
      "date": "2023-02-24T00:00:00-05:00",
      "open": "249.96",
      "high": "251.0",
      "low": "248.1",
      "close": "249.22",
      "volume": "24990905"
    },
    ...
  ]
}
```

## Implementation

Some notes on implementation.

- A very basic caching system is used, where the service stores the most recent response data in a map. When a request is made to the API the cache is checked first - if the cache is populated then the timestamp on the cached response is checked and returned if the results are less than 10 minutes old.
  - The alphavantage API allows for up to 500 requests per day, or ~20 requests per hour. To be safe, the cache expires after 10 minutes, or 6 request per hour per replica (3 replicas by default).
- This task is implemented with minimum third party libs.
  - `github.com/shopspring/decimal` provides arbitrary-precision fixed point decimal numbers, to avoid any floating point arithmetic shenanigans.
  - `golang.org/x/exp/slog` for logging. This is an experimental library from the Go team that could eventually make it into the stdlib. I haven't had a chance to play with it yet and thought this was a good time to take it for a test drive.
  - All other functionality is provided via the stdlib.