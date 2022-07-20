# h3server

## Name

*h3server* is a very simple plugin that starts an HTTP/3 web server that serves responses from a directory and gzips them using the same session keys and certificate as the DoQ server


## Syntax

~~~ txt
h3server DIRECTORY HOST:PORT
~~~

## Examples

Here's a very simple example:

~~~ txt
quic://.:784 {
	tls certs/example.crt certs/example.key {
		session_ticket_key test_session_ticket.key
	}
  
  ...
  
	h3server ./www/ localhost:4433
}
~~~

