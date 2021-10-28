openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca.key -out ca.pem -subj "/C=US/ST=Texas/L=Austin/OU=Tech/CN=ca"

openssl req -newkey rsa:4096 -nodes -keyout server.key -out server-req.pem -subj "/C=US/ST=Texas/L=Austin/OU=Tech/CN=server"

openssl x509 -req -in server-req.pem -days 60 -CA ca.pem -CAkey ca.key -CAcreateserial -out server.pem -extfile server.cnf

openssl req -newkey rsa:4096 -nodes -keyout client.key -out client-req.pem -subj "/C=US/ST=Texas/L=Austin/OU=Tech/CN=client"

openssl x509 -req -in client-req.pem -days 60 -CA ca.pem -CAkey ca.key -CAcreateserial -out client.pem -extfile client.cnf
