import struct

# Protocol wire format:
#   Request  (client → server): [2 bytes uint16 BE: payload length][payload: UTF-8 CSV string]
#   Response (server → client): [1 byte: 0x00=OK, 0x01=ERROR]

_ACK_OK = b'\x00'
_ACK_ERROR = b'\x01'


def recv_fields(sock) -> list[str]:
    """
    Read one CSV-encoded message from sock.
    Returns a list of string fields.
    Raises ConnectionError or ValueError on failure.
    """
    header = _recv_all(sock, 2)
    length = struct.unpack('!H', header)[0]

    payload = _recv_all(sock, length).decode('utf-8')
    return payload.split(',')


def recv_batch(sock) -> list[list[str]]:
    """
    Read a batch of CSV-encoded records from sock.
    Returns a list of field lists, one per record.
    Raises ConnectionError on EOF.
    """
    header = _recv_all(sock, 2)
    length = struct.unpack('!H', header)[0]

    payload = _recv_all(sock, length).decode('utf-8')
    rows = [row for row in payload.split('\n') if row]
    return [row.split(',') for row in rows]


def send_ack(sock, success: bool):
    """Send a 1-byte ACK to the client (0x00=OK, 0x01=ERROR)."""
    sock.sendall(_ACK_OK if success else _ACK_ERROR)


def _recv_all(sock, n: int) -> bytes:
    """Read exactly n bytes from sock, retrying on short reads."""
    data = b''
    while len(data) < n:
        chunk = sock.recv(n - len(data))
        if not chunk:
            raise ConnectionError(f"Connection closed after reading {len(data)}/{n} bytes")
        data += chunk
    return data


