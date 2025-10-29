# AES Encryption in Go

Today I implemented AES (Advanced Encryption Standard) encryption and decryption in Go using the `crypto/aes` and `crypto/cipher` packages. The implementation uses AES in CTR (Counter) mode.

## What is AES Encryption?

AES is a symmetric encryption algorithm, meaning the same key is used for both encryption and decryption. It is a block cipher, which means it operates on fixed-size blocks of data. The block size for AES is 128 bits (16 bytes). The key size can be 128, 192, or 256 bits. My implementation uses a 256-bit key (32 bytes), which is generated using `crypto/rand`.

## What is an IV?

IV stands for Initialization Vector. It is a random or pseudo-random value that is used to ensure that the same plaintext, when encrypted multiple times with the same key, results in a different ciphertext each time. This is important for security, as it prevents an attacker from recognizing patterns in the encrypted data.

In my implementation, the IV is generated using `crypto/rand` and is the same size as the AES block size (16 bytes). For encryption, the IV is prepended to the ciphertext. For decryption, the IV is read from the beginning of the ciphertext.

## How it Works

The `copyEncrypt` function takes a key, a source reader, and a destination writer. It first creates a new AES cipher with the given key. Then, it generates a random IV and writes it to the destination. Finally, it creates a new CTR stream cipher with the AES block and IV, and then reads from the source, encrypts the data, and writes it to the destination.

The `copyDecrypt` function takes a key, a source reader, and a destination writer. It also creates a new AES cipher with the given key. It then reads the IV from the beginning of the source. Finally, it creates a new CTR stream cipher with the AES block and IV, and then reads the encrypted data from the source, decrypts it, and writes it to the destination.

## How Secure is it?

AES with a 256-bit key is considered to be very secure. It is the standard for encryption used by the U.S. government and is widely used in the private sector as well. The security of the implementation depends on the secrecy of the key. As long as the key is kept secret, it is computationally infeasible to break the encryption.

The use of a random IV for each encryption ensures that the same plaintext will produce a different ciphertext each time it is encrypted with the same key, which adds another layer of security. The CTR mode of operation is also a secure and widely used mode.

The encypted file will be 16 bytes more than the actual file. The additional 16 bytes are for the Initialization Vector (IV), which is prepended
to the ciphertext.
