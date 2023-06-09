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
    # Host configuration
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.bind(("10.0.0.171", 9999))
    server.listen()

    # Accept connection from peer
    client, _ = server.accept()

    # Sending the host's public key to the peer
    client.send(public_key.save_pkcs1("PEM"))
    # Receiving the peer's public key
    public_partner = rsa.PublicKey.load_pkcs1(client.recv(1024))

    # Generate random AES key to use for communication (128, 192, or 256 bits long)
    aes_key = rsa.randnum.read_random_bits(128)
    print("AES KEY:", aes_key)

    # Encrypt the AES key with peer's public key
    aes_key_encrypted = rsa.encrypt(aes_key, public_partner)
    # Create digital signature for the encrypted AES key
    signature = rsa.sign(aes_key_encrypted, private_key, "SHA-256")

    # Send peer the encrypted AES key and digital signature
    client.send(aes_key_encrypted)
    client.send(signature)

elif choice == "2":
    # Peer configuration
    client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client.connect(("10.0.0.171", 9999))
    
    # Receiving the host's public key
    public_partner = rsa.PublicKey.load_pkcs1(client.recv(1024))

    # Send the peer's public key to host
    client.send(public_key.save_pkcs1("PEM"))

    # Receiving the encrypted AES key and digital signature from host
    aes_key_encrypted= client.recv(1024)
    host_signature = client.recv(1024)

    # Verify the digital signature of the encrypted AES key
    if rsa.verify(aes_key_encrypted, host_signature, public_partner):
        # Decrypt the AES key
        aes_key = rsa.decrypt(aes_key_encrypted, private_key)
        print("AES Key received and verified successfully.")
    else:
        # Verification failed so ending communication
        print("Failed to verify the AES Key's signature. Aborting.")
        exit()

    print("AES KEY:", aes_key)
else:
    exit()

def sending_messages(client, aes_key):
    while True:
        # Get input from host/peer to send
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
        # Generate a digital signature for the encrypted message
        message_signature = rsa.sign(encrypted, private_key, "SHA-256")
        
        # Send the encrypted message & signature
        client.send(encrypted)
        client.send(message_signature)
        # Print the plaintext message sent
        print("You: " + message)

def receiving_messages(client, aes_key):
    while True:
        # Receive the encrypted message
        encrypted_message = client.recv(1024)
        # Extract the digital signature from the received data
        message_signature = client.recv(1024)
        
        # Verify the digital signature
        if rsa.verify(encrypted_message, message_signature, public_partner):
            # The digital signature is valid, proceed with decryption and printing
            print("Message received and verified successfully.")
            # Extract the IV from the encrypted message
            iv = encrypted_message[:AES.block_size]
            # Create an AES cipher object with the key and mode (CBC)
            cipher = AES.new(aes_key, AES.MODE_CBC, iv)
            # Decrypt the message
            decrypted_message = cipher.decrypt(encrypted_message[AES.block_size:])
            # Remove the padding from the decrypted message
            unpadded_message = unpad(decrypted_message, AES.block_size)
            # Print out the decrypted received message
            print("Partner: ", unpadded_message.decode("utf-8"))
        else:
            # The digital signature is invalid, handle accordingly (end communication)
            print("Failed to verify the partner's signature. Aborting.")
            exit()


threading.Thread(target=sending_messages, args=(client, aes_key,)).start()
threading.Thread(target=receiving_messages, args=(client, aes_key,)).start()