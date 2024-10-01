
# CtrlB Collector

System of agents interacting with all-father server for telemetry collection and projection based on the Op-Amp specification.

Currently, it only supports fluent-bit. The project is built on a customized Fluent-bit library from [ctrlb-fluent-bit](https://github.com/ctrlb-hq/ctrlb-fluent-bit). Agent spins off a `fluent-bit` instance through its C-API via the C-Go interface and interacts for all the processes.

## Installation

Clone the GitHub repository

```bash
    git clone https://github.com/ctrlb-hq/ctrlb-collector.git
    cd ctrlb-collector
```

Build the docker container
```bash
    docker build -t ctrlb-collector .
```

Run the docker container
```bash
    docker run -it --network host ctrlb-collector
```

## API Reference

#### Start the agent instance

```http
POST /api/v1/start
```

#### Stop the agent instance

```http
POST /api/v1/stop
```
#### Retrieve the current config

```http
GET /api/v1/config
```

#### Update the current config

```http
PUT /api/v1/config
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `body`      | `string` | **Required**. new config json string |


#### Shutdown the agent instance

```http
POST /api/v1/shutdown
```



## Example Config
Following is an example fluent-bit config file in the `.json` format used in API calls, equivalent to the default `.yaml` [config file](https://github.com/ctrlb-hq/ctrlb-collector/blob/main/config.yaml).

```json
{
    "pipeline": {
        "filters": null,
        "inputs": [
            {
                "Interval_sec": 2,
                "name": "dummy"
            }
        ],
        "outputs": [
            {
                "match": "*",
                "name": "stdout"
            }
        ]
    },
    "service": {
        "http_server": "on"
    }
}
```
**NB**: The HTTP server in a fluent-bit instance is by default always set for metrics logging and runs with the following config:
```json
    "service": {
        "http_listen": "0.0.0.0",
        "http_port": 2020,
        "http_server": "on"
    }

```