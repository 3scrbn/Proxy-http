A small program acting as an HTTPS proxy that forwards incoming requests to the desired backend. It is useful in scenarios where exposing the backend is not desirable, and can be combined with firewall rules and other defensive measures.

The proxy includes a control parameter that prevents the request from being forwarded if the parameter is incorrect. This is useful for use in C2 (Command and Control) or similar setups.
