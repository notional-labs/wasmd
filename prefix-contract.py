import binascii
import bech32

# Define the prefixes similar to the Go code
CONTRACT_KEY_PREFIX = b'\x03'  # Corresponds to ContractKeyPrefix in Go

contract_addr = "wasm1896qu2ml7ckj675ukxjhx8xn0srp4e7cdx0m6rh949q9vw88hdrqj49ypr"
_, contract_data = bech32.bech32_decode(contract_addr.strip())  # Extract only the data part and strip any leading space

if isinstance(contract_data, list):  # Ensure it's the correct type
    contract_bytes = bech32.convertbits(contract_data, 5, 8, False)  # Convert from 5-bit to 8-bit
    if contract_bytes:
        contract_bytes = bytes(contract_bytes)
        contract_address_key = CONTRACT_KEY_PREFIX + contract_bytes
        print("Contract Address Key (hex):", binascii.hexlify(contract_address_key).decode())
    else:
        print("Error: Failed to convert address bits.")
else:
    print("Error: Decoded data is not in the expected format.")