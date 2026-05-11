import encodings.idna

tests = ["例", "例子", "日本", "日本語", "München"]

print("=== Python encodings.idna ===")
for t in tests:
    try:
        encoded = t.encode('idna')
        decoded = encoded.decode('idna')
        print(f"{t!r} -> {encoded!r}")
        print(f"  解码回: {decoded!r}")
    except Exception as e:
        print(f"{t!r}: Error: {e}")
