import binascii
import bech32

# Define the prefixes similar to the Go code
CONTRACT_KEY_PREFIX = b'\x02'  # Corresponds to ContractKeyPrefix in Go

# Additional components
PREFIX_0005 = b'\x00\x00\x05'  # Byte representation of 0x0005
STATE_BYTES = b'state'  # Byte representation of "state"

# Example user address (this should be replaced with the actual user address)
user_addr = "wasm1rvvlzguc0e2p8zpkq3ny28pvf5jdn5h6cfhmne"

# Decode the contract address
contract_addr = "wasm1r2eq2pkslppazxm2ks05zw35m7gykjmf3wtwtk364yp6q6cpc8zq52wcrx"
_, contract_data = bech32.bech32_decode(contract_addr.strip())

# Decode the user address
_, user_data = bech32.bech32_decode(user_addr.strip())

if isinstance(contract_data, list) and isinstance(user_data, list):
    contract_bytes = bech32.convertbits(contract_data, 5, 8, True)
    user_bytes = bech32.convertbits(user_data, 5, 8, True)
    
    if contract_bytes and user_bytes:
        contract_bytes = bytes(contract_bytes)
        user_bytes = bytes(user_bytes)
        
        # Combine all parts to form the full key
        full_key = CONTRACT_KEY_PREFIX + contract_bytes + PREFIX_0005 + STATE_BYTES + user_bytes
        print("Full Key (hex):", binascii.hexlify(full_key).decode())
    else:
        print("Error: Failed to convert address bits.")
else:
    print("Error: Decoded data is not in the expected format.")