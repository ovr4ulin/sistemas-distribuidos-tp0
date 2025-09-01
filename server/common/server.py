import socket
import logging
import signal


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._active = True
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(("", port))
        self._server_socket.listen(listen_backlog)

        # Setup signal handler
        signal.signal(signal.SIGTERM, self.__sigterm_handler)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        logging.info(f"action: run | result: in_progress")

        while self._active:
            client_sock = self.__accept_new_connection()
            if client_sock:
                self.__handle_client_connection(client_sock)

        logging.info(f"action: run | result: finished")

    def __sigterm_handler(self, signum, stack_frame):
        if not self._active:
            return

        self._active = False

        try:
            logging.info(f"action: close socket | result: in_progress")
            self._server_socket.close()
            logging.info(f"action: close socket | result: success")
        except OSError as e:
            logging.error(f"action: close socket | result: fail | error: {e}")

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # TODO: Modify the receive to avoid short-reads
            msg = client_sock.recv(1024).rstrip().decode("utf-8")
            addr = client_sock.getpeername()
            logging.info(
                f"action: receive_message | result: success | ip: {addr[0]} | msg: {msg}"
            )
            # TODO: Modify the send to avoid short-writes
            client_sock.send("{}\n".format(msg).encode("utf-8"))
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        try:
            logging.info("action: accept_connections | result: in_progress")
            c, addr = self._server_socket.accept()
            logging.info(
                f"action: accept_connections | result: success | ip: {addr[0]}"
            )
            return c
        except OSError as e:
            if self._active:
                logging.error(f"action: accept_connections | result: fail | error: {e}")
            else:
                logging.info(f"action: accept_connections | result: canceled")
            return None
