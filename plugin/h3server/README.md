# httpproxy

## Name

*httpproxy* is a very simple plugin that determines what to do with DOH HTTP requests that go to a location different from `/dns-query`.

## Description

The problem with DOH is that you usually would like to serve something useful from root and not just return 404. The `httpproxy` plugin is supposed to solve that problem. You just specify `host:port` where to proxy all requests that aren't served by the DOH server. 

## Syntax

~~~ txt
httpproxy HOST:PORT
~~~

## Examples

Here's a very simple example:

~~~ txt
httpproxy 127.0.0.1:8080
~~~

