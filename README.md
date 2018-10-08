# mackerel-plugin-request-pressure

## Synopsis

```console
% mackerel-plugin-request-pressure url
```

## Example of mackerel-agent.conf

```conf
[plugin.metrics.request-pressure-hoge]
command = "/path/to/mackerel-plugin-request-pressure -metric-key-prefix Top 'https://hogehoge.com'"
```
