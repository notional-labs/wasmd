import binascii
import bech32

# Define the prefixes similar to the Go code
CONTRACT_KEY_PREFIX = b'\x03'  # Corresponds to ContractKeyPrefix in Go
contract_addr = "wasm1dwl0wexpclhcu3r3zv0kvz0ggs2rk4a5svepaxqpmw2zmc9l42aqhfjwnp"
user_addr = "wasm1hj5fveer5cjtn4wd6wstzugjfdxzl0xpvsr89g"

## Get contract store path
_, contract_data = bech32.bech32_decode(contract_addr.strip())
contract_store_key_string = '' # string 
if isinstance(contract_data, list):  # Ensure it's the correct type
    contract_bytes = bech32.convertbits(contract_data, 5, 8, False)  # Convert to 8-bit bytes
    if contract_bytes:
        contract_bytes = bytes(contract_bytes)
        contract_address_key_raw = CONTRACT_KEY_PREFIX + contract_bytes
        contract_store_key_string = binascii.hexlify(contract_address_key_raw).decode()  # Keep as bytes
    else:
        print("Error: Failed to convert address bits.")
else:
    print("Error: Decoded data is not in the expected format.")

# Get counter store path 
PREFIX_0005 = b'\x00\x05'  # 5 len represents length of namespace "state"
STATE_BYTES = b'state'  # Byte representation of "state"

# Encode the user address to bytes
user_bytes = user_addr.encode()
counter_key =  PREFIX_0005 + STATE_BYTES + user_bytes

# Print the full key in hexadecimal format
counter_key_string = binascii.hexlify(counter_key).decode()
full_key = contract_store_key_string + counter_key_string
print("Full Key (hex):", full_key)

