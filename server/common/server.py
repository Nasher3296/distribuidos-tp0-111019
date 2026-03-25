import signal
import socket
import logging
import threading

from common.protocol import (
    recv_message, send_ack, send_message,
    MSG_TYPE_BATCH, MSG_TYPE_DONE, MSG_TYPE_QUERY_WINNERS, MSG_TYPE_WINNERS,
    RECORD_SEPARATOR, FIELD_SEPARATOR,
)
from common.utils import bet_from_fields, store_bets, load_bets, has_won

class Server:
    def __init__(self, port, listen_backlog, number_of_agencies):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        self._number_of_agencies = number_of_agencies
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

        self._store_lock = threading.Lock()
        self._agencies_done = 0
        self._agencies_lock = threading.Lock()
        self._lottery_done = threading.Event()
        self._winners = {}

    def run(self):
        threads = []
        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                t = threading.Thread(target=self.__handle_client_connection, args=(client_sock,), daemon=True)
                t.start()
                threads.append(t)
            except OSError as e:
                if self._running:
                    logging.error(f"action: accept_connections | result: fail | error: {e}")
                break

        for t in threads:
            t.join()
        logging.info("action: stop_server | result: success")

    def __handle_sigterm(self, *args):
        logging.info("action: sigterm_received | result: success")
        self._running = False
        self._server_socket.close()
        logging.info("action: close_server_socket | result: success")

    def __handle_client_connection(self, client_sock):
        addr = client_sock.getpeername()
        agency_id = None
        try:
            while self._running:
                msg_type, payload = recv_message(client_sock)

                if msg_type == MSG_TYPE_BATCH:
                    rows = [r for r in payload.decode('utf-8').split(RECORD_SEPARATOR) if r]
                    batch = [row.split(FIELD_SEPARATOR) for row in rows]
                    bets = [bet_from_fields(fields) for fields in batch]
                    with self._store_lock:
                        store_bets(bets)
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    send_ack(client_sock, True)

                elif msg_type == MSG_TYPE_DONE:
                    agency_id = int(payload.decode('utf-8'))
                    self.__notify_done()
                    send_ack(client_sock, True)
                    break

            self._lottery_done.wait()

            msg_type, payload = recv_message(client_sock)
            if msg_type == MSG_TYPE_QUERY_WINNERS:
                winners = self._winners.get(agency_id, [])
                response = RECORD_SEPARATOR.join(winners).encode('utf-8')
                send_message(client_sock, MSG_TYPE_WINNERS, response)

        except Exception as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            try:
                send_ack(client_sock, False)
            except Exception:
                pass
        finally:
            client_sock.close()
            logging.info(f'action: close_connection | result: success | ip: {addr[0]}')

    def __notify_done(self):
        """Increment done counter; when all agencies are done, run the lottery."""
        with self._agencies_lock:
            self._agencies_done += 1
            if self._agencies_done == self._number_of_agencies:
                self.__run_lottery()
                self._lottery_done.set()

    def __run_lottery(self):
        winners = {}
        for bet in load_bets():
            if has_won(bet):
                winners.setdefault(bet.agency, []).append(bet.document)
        self._winners = winners
        logging.info("action: sorteo | result: success")

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
