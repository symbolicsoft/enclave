# Enclave

**Enclave** is your command-line secure encrypted deniable cloud-synchronized notebook.

Enclaves can be accessed from anywhere with a simple, memorizable ten-word passphrase (without an accompanying username) and benefit from blazing-fast synchronization. Enclave notebooks are authenticated yet anonymous, provide plausible deniability and self-destruction features, and are highly portable thanks to their passphrase access system, which in practice works in a way similar to cryptocurrency wallets.

Enclave notebooks benefit from state-of-the art post-quantum secure encryption. The Enclave command-line notebook editor is simple yet full-featured, and works on all major operating systems.

Each Enclave notebook can be set up to work with an optional _decoy_ notebook: when a decoy notebook is accessed, its paired notebook is wiped from the Enclave server. It is not possible for an attacker who does not have access to the server to determine if any Enclave notebook is a decoy notebook or if it is paired to a decoy notebook, or if an Enclave passphrase points to a wiped notebook.

## Technical Specification

Enclave Protocol is meant to provide highly portable secure notebook synchronization from a light client.

### Principals

- **Alice**: a local client user who owns a notebook.
- **Server**: a remote notebook synchronization server.

### Security Properties

- **Confidentiality**: Notebook contents are only visible to Alice.
- **Authentication**: Notebook contents cannot be undetectably modified by any party other than Alice.
- **Anonymity**: Enclave notebooks are not tied down to any specific user identifier: the only identifier is a randomly generated passphrase, much like cryptocurrency wallets.
- **Deniability and self-destruction**: Alice can access different notebooks that quietly wipe paired notebooks based on the provided key material.
- **Portability**: Notebook access must be predicated on singular, portable (human-readable, memorizable) key material.

### Key Generation

Alice's key generation flow looks like this:

```text
 Mandatory:
+----------+
|  User    | PUS = SCRYPT(US, salt, N=2^20, r=8, p=1)
|  Secret  -------------------------------------->--------+
|  (US)    | BLAKE2X(PUS)                      0 | USK-ID |
+----------+                                     |--------+
                                               1 | USK-ED |
                                                 +--------+
 Optional:
+----------+
|  Decoy   | PDS = SCRYPT(DS, salt, N=2^20, r=8, p=1)
|  Secret  -------------------------------------->--------+
|  (DS)    | BLAKE2X(PDS)                      0 | USK-DD |
+----------+                                     |--------+
                                               1 | USK-DX |
                                                 +--------+
```

- `US`: 12-word mnemonic chosen randomly out of a list of 5459 words.
- `DS`: 12-word mnemonic chosen randomly out of a list of 5459 words.
- `salt`: the byte representation of the 24-byte string `DTWdTA8L9VZG5J8p5dNaUmrQ`.
- `USK-ID`: a string used to identify her notebook to the server.
- `USK-ED`: the notebook 256-bit encryption key.
- `USK-DD`: identifier string, but for the decoy notebook.
- `USK-DX`: decoy notebook 256-bit encryption key.

#### Note on Key Reuse

Whenever Alice updates her notebook, she will be using the same `USK-ED` to re-encrypt it. As such, in order to avoid nonce reuse, it becomes crucial to use an extended nonce cipher, which is why we use `XChaCha20-Poly1305`, which employs 192-bit nonces.

With 192-bit nonces, the chance of nonce reuse for 100,000,000 encryptions may be estimated as `(2^192)/(10^8) ~= 2^169`. These are acceptable numbers, so we can proceed with the chosen key, cipher and nonce size.

#### Note on Key Enumeration

The key space passphrases is `WordlistSize^PassphraseLength = 5459^12 = 2^149`. These passphrases are run through an expensive Scrypt operation to produce 256-bit hashes, which are then used to derive more 256-bit subkeys.

There is no realistic risk for key collision on the 256-bit hashes or subkeys. However, because all Scrypt hashes are produced with a static salt, an increase in the number of encrypted notebooks means a theoretical increase in the possibility for enumerating a passphrase for some random existing encrypted notebook: if a server has one encrypted notebook, the chance of guessing a random encrypted notebook is `(2^149)/1 = 2^149`. If a server has 100,000,000 encrypted notebooks, that chance becomes `(2^149)/(10^8) ~= 2^122`. Especially given the very high cost of enumerating the passphrase hash space with the chosen Scrypt parameters, these are acceptable numbers, so we can proceed with the chosen passphrase size.

### Transport Layer

Enclave uses [gRPC](https://grpc.io) as the transport layer, chosen for its speed and efficiency. Transport layer authentication is guaranteed by hardcoding the server's X.509 elliptic-curve public key within the Enclave client.

### Storage & Synchronization

#### Alice

Alice stores:

- `(USK-ID, USK-ED)` **or** `(USK-DD, USK-DX)`.
  - Storing either subkey tuple is completely optional.
  - Alice cannot store both keys simultaneously.

Alice can set a small "PIN" used to encrypt stored subkeys. Given that users will have a preference towards this PIN being short and easy to quickly type, we'll be applying Scrypt again (with high cost parameters) when generating the subkey encryption keys from it:

1. Alice provides `PIN`.
2. Enclave generates a random 24-bit `salt`.
3. Enclave calculates `CEK = SCRYPT(PIN, salt, N=2^20, r=8, p=1)`
4. Enclave encrypts locally stored subkeys with CEK using XChaCha20-Poly1305 and a random nonce.

#### Server

Server stores:

- Alice's notebook `NR` under her `USK-ID`.
- Alice's decoy notebook `ND` under her `USK-DD` (optional).
- Alice's last-used encryption nonce.

### User Flow

#### First Run

From Alice's perspective:

1. Alice runs `enclave` for the first time.
2. `enclave` asks Alice if she'd like to set up a new notebook. Alice says yes.
3. `enclave` checks if it's able to open a connection to `enclave-server` and aborts if not.
4. `enclave` generates `US` and communicates it to Alice.
5. `enclave` generates `USK-ID` and sends it to `enclave-server`.
6. `enclave` asks Alice if she'd like to set up a decoy notebook. If Alice accepts:
    - `enclave` advises Alice that she can quickly generate one using ChatGPT, providing example prompts.
    - `enclave` generates `DS` and communicates it to Alice.
    - `enclave` generates `USK-DD` and sends it to `enclave-server` along with notebook `DS` encrypted with `USK-DX`.
    - `enclave` communicates `DS` to Alice.

From Server's perspective:

1. Server receives a request to store a tuple of notebooks (`NR`, `ND`) under their respective identifiers `USK-ID` and `USK-DD`.
    - `ND` and `USK-DD` are optional.
2. `NR` is stored under the identifier `USK-ID`.
    - Should it exist, `ND` is stored under the identifier `USK-DD`.

#### Subsequent Runs

Upon Alice running `enclave`:

1. `enclave` checks if `US` is stored locally. If it is, we can quickly fetch and decrypt the notebook from Server.
    - `US` could potentially be a decoy secret (`DS`).
2. If `US` is not stored locally, Alice is asked to input her mnemonic. At this point, she may input `US` or (should it exist) `DS`.

From Server's perspective:

1. Whenever anyone requests the notebook with identifier `USK-ID`, Server sends `NR` along with its stored last-used encryption nonce.
2. Whenever anyone requests the notebook with identifier `USK-DD`, Server **deletes** `USK-ID` and `NR` (if not already deleted) and sends `ND`.
3. For any real or decoy notebook identifier that does not exist or has been deleted, Server responds with a "notebook not found" error.

### Restrictions

- 64 pages per notebook.
- 64KB per notebook page.

## Author

Written by Nadim Kobeissi (Symbolic Software), released under the GNU GPLv2 license.
