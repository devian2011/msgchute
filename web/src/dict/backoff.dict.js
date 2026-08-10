export var backOffParams = {
    linear: {duration: "2s"},
    "jitter-linear": {duration: "2s", jitter: 0.2},
    exponential: {multiplier: 2.0, baseDelay: "1s", maxDelay: "5m"},
    "jitter-exponential":  {multiplier: 2.0, baseDelay: "1s", maxDelay: "5m", jitter: "0.2"},
};