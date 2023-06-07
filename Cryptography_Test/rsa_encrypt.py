import socket
import threading
import rsa

public_key, private_key = rsa.newkeys(1024)
public_partner = None

# Get input to see whether they will host or connect in p2p connection
choice = input("Do you want to host (1) or connect (2)?")

if choice == "1":
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind(("10.0.0.171", 9999))
    server.listen()

    client, _ = server.accept()
    client.send(public_key.save_pkcs1("PEM"))
    public_partner = rsa.PublicKey.load_pkcs1(client.recv(1024))

elif choice == "2":
    client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client.connect(("10.0.0.171", 9999))
    public_partner = rsa.PublicKey.load_pkcs1(client.recv(1024))
    client.send(public_key.save_pkcs1("PEM"))

else:
    exit()

def sending_messages(client):
    while True:
        message = input("")

        client.send(rsa.encrypt(message.encode(), public_partner))
        print("You: " + message)

def receiving_messages(client):
    while True:
        print("Partner: " + rsa.decrypt(client.recv(1024), private_key).decode())

threading.Thread(target=sending_messages, args=(client,)).start()
threading.Thread(target=receiving_messages, args=(client,)).start()