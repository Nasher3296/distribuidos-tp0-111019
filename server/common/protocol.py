import struct

MSG_TYPE_BATCH         = 0x01
MSG_TYPE_DONE          = 0x02
MSG_TYPE_QUERY_WINNERS = 0x03

MSG_TYPE_ACK_OK        = 0x10
MSG_TYPE_ACK_ERROR     = 0x11
MSG_TYPE_WINNERS       = 0x12

RECORD_SEPARATOR = '\n'
FIELD_SEPARATOR  = ','

#   [1 byte: message type][2 bytes uint16 BE: payload length][payload: UTF-8]


def recv_message(sock) -> tuple[int, bytes]:
    """
    Read one canonical message from sock.
    Returns (msg_type, raw_payload).
    Raises ConnectionError on EOF.
    """
    type_byte = _recv_all(sock, 1)[0]
    header = _recv_all(sock, 2)
    length = struct.unpack('!H', header)[0]
    payload = _recv_all(sock, length) if length > 0 else b''
    return type_byte, payload


def recv_batch(sock) -> list[list[str]]:
    """
    Read a batch of CSV-encoded records from sock.
    Returns a list of field lists, one per record.
    Raises ConnectionError on EOF.
    """
    _, payload = recv_message(sock)
    rows = [row for row in payload.decode('utf-8').split(RECORD_SEPARATOR) if row]
    return [row.split(FIELD_SEPARATOR) for row in rows]


def send_ack(sock, success: bool):
    """Send an ACK response in canonical format (empty payload)."""
    msg_type = MSG_TYPE_ACK_OK if success else MSG_TYPE_ACK_ERROR
    sock.sendall(bytes([msg_type]) + struct.pack('!H', 0))


def _recv_all(sock, n: int) -> bytes:
    """Read exactly n bytes from sock, retrying on short reads."""
    data = b''
    while len(data) < n:
        chunk = sock.recv(n - len(data))
        if not chunk:
            raise ConnectionError(f"Connection closed after reading {len(data)}/{n} bytes")
        data += chunk
    return data
