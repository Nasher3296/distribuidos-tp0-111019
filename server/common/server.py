import signal
import socket
import logging

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
        self._winners = {}
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def run(self):
        waiting_sockets = []

        while self._running and len(waiting_sockets) < self._number_of_agencies:
            try:
                client_sock = self.__accept_new_connection()
            except OSError as e:
                if self._running:
                    logging.error(f"action: accept_connections | result: fail | error: {e}")
                break

            agency_id = self.__receive_bets(client_sock)
            if agency_id is not None:
                waiting_sockets.append((agency_id, client_sock))
            else:
                client_sock.close()

        if len(waiting_sockets) == self._number_of_agencies:
            self.__run_lottery()
            for agency_id, sock in waiting_sockets:
                self.__respond_to_query(agency_id, sock)
        else:
            for _, sock in waiting_sockets:
                sock.close()

        logging.info("action: stop_server | result: success")

    def __handle_sigterm(self, *args):
        logging.info("action: sigterm_received | result: success")
        self._running = False
        self._server_socket.close()
        logging.info("action: close_server_socket | result: success")

    def __receive_bets(self, client_sock):
        """Receive all bet batches from a client until DONE. Returns agency_id."""
        addr = client_sock.getpeername()
        agency_id = None
        try:
            while self._running:
                msg_type, payload = recv_message(client_sock)

                if msg_type == MSG_TYPE_BATCH:
                    rows = [r for r in payload.decode('utf-8').split(RECORD_SEPARATOR) if r]
                    batch = [row.split(FIELD_SEPARATOR) for row in rows]
                    bets = [bet_from_fields(fields) for fields in batch]
                    store_bets(bets)
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    send_ack(client_sock, True)

                elif msg_type == MSG_TYPE_DONE:
                    agency_id = int(payload.decode('utf-8'))
                    send_ack(client_sock, True)
                    break

        except Exception as e:
            logging.error(f"action: receive_bets | result: fail | ip: {addr[0]} | error: {e}")
            try:
                send_ack(client_sock, False)
            except Exception:
                pass

        return agency_id

    def __respond_to_query(self, agency_id, client_sock):
        """Wait for a QUERY_WINNERS message and respond with the winners list."""
        addr = client_sock.getpeername()
        try:
            msg_type, _ = recv_message(client_sock)
            if msg_type == MSG_TYPE_QUERY_WINNERS:
                winners = self._winners.get(agency_id, [])
                response = RECORD_SEPARATOR.join(winners).encode('utf-8')
                send_message(client_sock, MSG_TYPE_WINNERS, response)
                logging.info(f'action: winners_sent | result: success | agency_id: {agency_id} | cant_ganadores: {len(winners)}')
        except Exception as e:
            logging.error(f"action: respond_to_query | result: fail | ip: {addr[0]} | error: {e}")
        finally:
            client_sock.close()
            logging.info(f'action: close_connection | result: success | ip: {addr[0]}')

    def __run_lottery(self):
        for bet in load_bets():
            if has_won(bet):
                self._winners.setdefault(bet.agency, []).append(bet.document)
        logging.info("action: sorteo | result: success")

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
