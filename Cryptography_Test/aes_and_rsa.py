import socket
import threading
import rsa
from Crypto.Cipher import AES
from Crypto.Random import get_random_bytes
from Crypto.Util.Padding import pad, unpad

public_key, private_key = rsa.newkeys(1024)
public_partner = None

aes_key = None

# Get input to see whether they will host or connect in p2p connection
choice = input("Do you want to host (1) or connect (2)?")

if choice == "1":
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind(("10.0.0.171", 9999))
    server.listen()

    client, _ = server.accept()
    client.send(public_key.save_pkcs1("PEM"))
    public_partner = rsa.PublicKey.load_pkcs1(client.recv(1024))

    aes_key = rsa.randnum.read_random_bits(128)
    print(aes_key)
    client.send(rsa.encrypt(aes_key, public_partner))
elif choice == "2":
    client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client.connect(("10.0.0.171", 9999))
    public_partner = rsa.PublicKey.load_pkcs1(client.recv(1024))
    client.send(public_key.save_pkcs1("PEM"))

    aes_key = rsa.decrypt(client.recv(1024), private_key)
    print(aes_key)
else:
    exit()

def sending_messages(client, aes_key):
    while True:
        message = input("")

        # Generate a random initialization vector (IV)
        iv = get_random_bytes(AES.block_size)
        # Create an AES cipher object with the key and mode (CBC)
        cipher = AES.new(aes_key, AES.MODE_CBC, iv)
        # Pad the message to match the block size
        padded_message = pad(message.encode(), AES.block_size)
        # Encrypt the padded message
        encrypted_message = cipher.encrypt(padded_message)
        # Encrypt the message
        encrypted = iv + encrypted_message

        client.send(encrypted)
        print("You: " + message)

def receiving_messages(client, aes_key):
    while True:
        encrypted_message = client.recv(1024)
        # Extract the IV from the encrypted message
        iv = encrypted_message[:AES.block_size]
        # Create an AES cipher object with the key and mode (CBC)
        cipher = AES.new(aes_key, AES.MODE_CBC, iv)
        # Decrypt the message
        decrypted_message = cipher.decrypt(encrypted_message[AES.block_size:])
        # Remove the padding from the decrypted message
        unpadded_message = unpad(decrypted_message, AES.block_size)
        print("Partner: ", unpadded_message.decode("utf-8"))

threading.Thread(target=sending_messages, args=(client, aes_key,)).start()
threading.Thread(target=receiving_messages, args=(client, aes_key,)).start()