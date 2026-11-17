"""独立复算 Java PasswordUtil.encrypt(username, password, salt) 的结果，
用于与 Go 端 LegacyPBE 实现交叉验证。

算法：PBEWithMD5AndDES（PKCS#5 v1.5）
  - PBKDF1(MD5) 派生：T1 = MD5(pwd || salt)，Ti = MD5(T(i-1))，共 1000 次
  - 取 16 字节：key = 前 8 字节，iv = 后 8 字节
  - DES-CBC 加密明文，PKCS#5 填充，输出 hex

DES 使用纯 Python 实现，避免额外依赖（本机无法安装 pycryptodome）。
"""
import hashlib
import struct
import sys

# --- 标准 DES 实现（仅用于交叉验证，非性能场景）---
IP = [58,50,42,34,26,18,10,2,60,52,44,36,28,20,12,4,
      62,54,46,38,30,22,14,6,64,56,48,40,32,24,16,8,
      57,49,41,33,25,17,9,1,59,51,43,35,27,19,11,3,
      61,53,45,37,29,21,13,5,63,55,47,39,31,23,15,7]
FP = [40,8,48,16,56,24,64,32,39,7,47,15,55,23,63,31,
      38,6,46,14,54,22,62,30,37,5,45,13,53,21,61,29,
      36,4,44,12,52,20,60,28,35,3,43,11,51,19,59,27,
      34,2,42,10,50,18,58,26,33,1,41,9,49,17,57,25]
E = [32,1,2,3,4,5,4,5,6,7,8,9,8,9,10,11,12,13,
     12,13,14,15,16,17,16,17,18,19,20,21,20,21,22,23,24,25,
     24,25,26,27,28,29,28,29,30,31,32,1]
P = [16,7,20,21,29,12,28,17,1,15,23,26,5,18,31,10,
     2,8,24,14,32,27,3,9,19,13,30,6,22,11,4,25]
PC1 = [57,49,41,33,25,17,9,1,58,50,42,34,26,18,
       10,2,59,51,43,35,27,19,11,3,60,52,44,36,
       63,55,47,39,31,23,15,7,62,54,46,38,30,22,
       14,6,61,53,45,37,29,21,13,5,28,20,12,4]
PC2 = [14,17,11,24,1,5,3,28,15,6,21,10,
       23,19,12,4,26,8,16,7,27,20,13,2,
       41,52,31,37,47,55,30,40,51,45,33,48,
       44,49,39,56,34,53,46,42,50,36,29,32]
SHIFTS = [1,1,2,2,2,2,2,2,1,2,2,2,2,2,2,1]
SBOX = [
 [14,4,13,1,2,15,11,8,3,10,6,12,5,9,0,7,0,15,7,4,14,2,13,1,10,6,12,11,9,5,3,8,
  4,1,14,8,13,6,2,11,15,12,9,7,3,10,5,0,15,12,8,2,4,9,1,7,5,11,3,14,10,0,6,13],
 [15,1,8,14,6,11,3,4,9,7,2,13,12,0,5,10,3,13,4,7,15,2,8,14,12,0,1,10,6,9,11,5,
  0,14,7,11,10,4,13,1,5,8,12,6,9,3,2,15,13,8,10,1,3,15,4,2,11,6,7,12,0,5,14,9],
 [10,0,9,14,6,3,15,5,1,13,12,7,11,4,2,8,13,7,0,9,3,4,6,10,2,8,5,14,12,11,15,1,
  13,6,4,9,8,15,3,0,11,1,2,12,5,10,14,7,1,10,13,0,6,9,8,7,4,15,14,3,11,5,2,12],
 [7,13,14,3,0,6,9,10,1,2,8,5,11,12,4,15,13,8,11,5,6,15,0,3,4,7,2,12,1,10,14,9,
  10,6,9,0,12,11,7,13,15,1,3,14,5,2,8,4,3,15,0,6,10,1,13,8,9,4,5,11,12,7,2,14],
 [2,12,4,1,7,10,11,6,8,5,3,15,13,0,14,9,14,11,2,12,4,7,13,1,5,0,15,10,3,9,8,6,
  4,2,1,11,10,13,7,8,15,9,12,5,6,3,0,14,11,8,12,7,1,14,2,13,6,15,0,9,10,4,5,3],
 [12,1,10,15,9,2,6,8,0,13,3,4,14,7,5,11,10,15,4,2,7,12,9,5,6,1,13,14,0,11,3,8,
  9,14,15,5,2,8,12,3,7,0,4,10,1,13,11,6,4,3,2,12,9,5,15,10,11,14,1,7,6,0,8,13],
 [4,11,2,14,15,0,8,13,3,12,9,7,5,10,6,1,13,0,11,7,4,9,1,10,14,3,5,12,2,15,8,6,
  1,4,11,13,12,3,7,14,10,15,6,8,0,5,9,2,6,11,13,8,1,4,10,7,9,5,0,15,14,2,3,12],
 [13,2,8,4,6,15,11,1,10,9,3,14,5,0,12,7,1,15,13,8,10,3,7,4,12,5,6,11,0,14,9,2,
  7,11,4,1,9,12,14,2,0,6,10,13,15,3,5,8,2,1,14,7,4,10,8,13,15,12,9,0,3,5,6,11],
]


def permute(block, table, in_bits):
    out = 0
    for pos in table:
        out = (out << 1) | ((block >> (in_bits - pos)) & 1)
    return out


def subkeys(key8):
    k = int.from_bytes(key8, "big")
    pc1 = permute(k, PC1, 64)
    c = (pc1 >> 28) & 0xFFFFFFF
    d = pc1 & 0xFFFFFFF
    keys = []
    for shift in SHIFTS:
        c = ((c << shift) | (c >> (28 - shift))) & 0xFFFFFFF
        d = ((d << shift) | (d >> (28 - shift))) & 0xFFFFFFF
        keys.append(permute((c << 28) | d, PC2, 56))
    return keys


def des_encrypt_block(block8, key8):
    keys = subkeys(key8)
    m = int.from_bytes(block8, "big")
    m = permute(m, IP, 64)
    l = (m >> 32) & 0xFFFFFFFF
    r = m & 0xFFFFFFFF
    for k in keys:
        er = permute(r, E, 32) ^ k
        s_out = 0
        for i in range(8):
            six = (er >> (42 - 6 * i)) & 0x3F
            row = ((six & 0x20) >> 4) | (six & 1)
            col = (six >> 1) & 0xF
            s_out = (s_out << 4) | SBOX[i][row * 16 + col]
        f = permute(s_out, P, 32)
        l, r = r, l ^ f
    pre = (r << 32) | l
    return permute(pre, FP, 64).to_bytes(8, "big")


def des_decrypt_block(block8, key8):
    keys = list(reversed(subkeys(key8)))
    m = permute(int.from_bytes(block8, "big"), IP, 64)
    l = (m >> 32) & 0xFFFFFFFF
    r = m & 0xFFFFFFFF
    for k in keys:
        er = permute(r, E, 32) ^ k
        s_out = 0
        for i in range(8):
            six = (er >> (42 - 6 * i)) & 0x3F
            row = ((six & 0x20) >> 4) | (six & 1)
            col = (six >> 1) & 0xF
            s_out = (s_out << 4) | SBOX[i][row * 16 + col]
        f = permute(s_out, P, 32)
        l, r = r, l ^ f
    pre = (r << 32) | l
    return permute(pre, FP, 64).to_bytes(8, "big")


def pbkdf1_md5(password: bytes, salt: bytes, iterations: int, dklen: int = 16) -> bytes:
    t = hashlib.md5(password + salt).digest()
    material = bytearray(t)
    cur = t
    for _ in range(1, iterations):
        cur = hashlib.md5(cur).digest()
        material.extend(cur)
    return bytes(material[:dklen])


def pkcs5_pad(data: bytes, block=8) -> bytes:
    pad = block - len(data) % block
    return data + bytes([pad]) * pad


def java_encrypt(username: str, password: str, salt: str) -> str:
    plaintext = username.encode("utf-8")
    key_material = pbkdf1_md5(password.encode("utf-8"), salt.encode("utf-8"), 1000, 16)
    key, iv = key_material[:8], key_material[8:16]
    data = pkcs5_pad(plaintext)
    out = b""
    prev = iv
    for i in range(0, len(data), 8):
        blk = bytes(a ^ b for a, b in zip(data[i:i+8], prev))
        enc = des_encrypt_block(blk, key)
        out += enc
        prev = enc
    return out.hex()


def java_decrypt(cipher_hex: str, password: str, salt: str) -> str:
    raw = bytes.fromhex(cipher_hex)
    key_material = pbkdf1_md5(password.encode("utf-8"), salt.encode("utf-8"), 1000, 16)
    key, iv = key_material[:8], key_material[8:16]
    out = b""
    prev = iv
    for i in range(0, len(raw), 8):
        dec = des_decrypt_block(raw[i:i+8], key)
        out += bytes(a ^ b for a, b in zip(dec, prev))
        prev = raw[i:i+8]
    pad = out[-1]
    return out[:-pad].decode("utf-8")


# 自检：FIPS 已知向量（DES ECB）
def selftest():
    pt = bytes.fromhex("0123456789ABCDEF")
    key = bytes.fromhex("133457799BBCDFF1")
    got = des_encrypt_block(pt, key).hex().upper()
    assert got == "85E813540F0AB405", got
    print("DES FIPS 向量自检通过:", got)


if __name__ == "__main__":
    selftest()
    cases = [
        ("admin", "123456", "63293188"),
        ("teacher", "123456", "63293188"),
        ("student", "Abc@12345", "63293188"),
        ("张三", "passw0rd", "63293188"),
    ]
    for u, p, s in cases:
        c = java_encrypt(u, p, s)
        back = java_decrypt(c, p, s)
        assert back == u, (back, u)
        print(f"{u}\t{p}\t{s}\t{c}")
