import binascii
import bech32

# Define the prefixes similar to the Go code
# Additional components
PREFIX_0005 = b'\x00\x00\x05'  # Byte representation of 0x0005
STATE_BYTES = b'state'  # Byte representation of "state"

# Example user address (this should be replaced with the actual user address)
user_addr = "wasm1sejq8wn2qyw0vhgk3y0ds3r6a7fssl57776aek"
# Decode the contract address

# Decode the user address
_, user_data = bech32.bech32_decode(user_addr.strip())

if isinstance(user_data, list):
    user_bytes = bech32.convertbits(user_data, 5, 8, True)
    
    if user_bytes:
        user_bytes = bytes(user_bytes)
        
        # Combine all parts to form the full key
        full_key = PREFIX_0005 + STATE_BYTES + user_bytes
        print("Full Key (hex):", binascii.hexlify(full_key).decode())
    else:
        print("Error: Failed to convert address bits.")
else:
    print("Error: Decoded data is not in the expected format.")