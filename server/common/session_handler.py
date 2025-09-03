import socket
from common.services import Service


class SessionHandler:
    def __init__(self, conn: socket.socket, service: Service) -> None:
        self.conn: socket.socket = conn
        self.service: Service = service

    def run(self) -> None:
        message_encoded: bytes = self._read_message_from_socket()
        response_encoded: bytes | None = self.service.handle_message(message_encoded)
        if response_encoded is not None:
            self._write_message_to_socket(response_encoded)

    def _read_message_from_socket(self) -> bytes:
        message_length_encoded: bytes = self._read_exact(4)
        message_length: int = int.from_bytes(
            message_length_encoded, byteorder="big", signed=False
        )
        message_encoded: bytes = self._read_exact(message_length)
        return message_encoded

    def _write_message_to_socket(self, message_encoded: bytes) -> None:
        message_length: int = len(message_encoded)
        message_length_encoded: bytes = int.to_bytes(
            message_length, length=4, byteorder="big", signed=False
        )
        self._write_exact(message_length_encoded)
        self._write_exact(message_encoded)

    def _read_exact(self, n: int) -> bytes:
        buf = b""

        while len(buf) < n:
            chunk = self.conn.recv(n - len(buf))
            if not chunk:
                raise ConnectionError("Socket cerrado antes de recibir los datos")
            buf += chunk

        return buf

    def _write_exact(self, data: bytes) -> None:
        total_sent = 0
        while total_sent < len(data):
            sent = self.conn.send(data[total_sent:])
            if sent == 0:
                raise ConnectionError("Socket cerrado antes de enviar todos los datos")
            total_sent += sent
