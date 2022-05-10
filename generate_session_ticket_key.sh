#!/bin/bash
#tls plugin takes last 32 Bytes
#the test session ticket has 48 Bytes so we just use the same amount
openssl rand -out session_ticket.key 48