import signal
import socket
import logging

from common.protocol import recv_fields, send_ack
from common.utils import bet_from_fields, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError as e:
                if not self._running:
                    logging.info("action: stop_server | result: success")
                else:
                    logging.error(f"action: accept_connections | result: fail | error: {e}")
                break

        logging.info("action: stop_server | result: success")

    def __handle_sigterm(self, *args):
        logging.info("action: sigterm_received | result: success")
        self._running = False
        self._server_socket.close()
        logging.info("action: close_server_socket | result: success")

    def __handle_client_connection(self, client_sock):
        """
        Receive a bet from the client, persist it, and send an ACK.

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        addr = client_sock.getpeername()
        try:
            fields = recv_fields(client_sock)
            bet = bet_from_fields(fields)
            store_bets([bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
            send_ack(client_sock, True)
        except Exception as e:
            logging.error(f"action: receive_bet | result: fail | error: {e}")
            try:
                send_ack(client_sock, False)
            except Exception:
                pass
        finally:
            client_sock.close()
            logging.info(f'action: close_connection | result: success | ip: {addr[0]}')

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
