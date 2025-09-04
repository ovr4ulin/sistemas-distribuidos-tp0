from common.messages import (
    BetMessage,
    AckMessage,
    Message,
    BetBatchMessage,
    EndOfBetsMessage,
    WinnersNotificationMessage,
)
from common.utils import Bet, store_bets, load_bets, has_won
from abc import ABC, abstractmethod
from typing import Optional
import os
import logging
from threading import RLock, Event

########################################################
# ABSTRACT SERVICE
########################################################


class Service(ABC):
    @abstractmethod
    def handle_message(self, message_encoded: bytes) -> Optional[bytes]:
        raise NotImplementedError


########################################################
# SERVICES
########################################################


class BetCoordinator:
    def __init__(self):
        self._winners_per_agency: dict[str, list[Bet]] = {}
        self._agencies_proccesed: set = set()
        self._total_agencies: int = int(os.environ.get("TOTAL_AGENCIES"))
        self._winners_ready: Event = Event()
        self._lock: RLock = RLock()

    def store_bets(self, bets: list[Bet]) -> None:
        with self._lock:
            store_bets(bets)

    def mark_end_of_bets(self, agency: str) -> None:
        with self._lock:
            self._agencies_proccesed.add(agency)

            if len(self._agencies_proccesed) >= self._total_agencies:
                self._calculate_winners()
                self._winners_ready.set()

    def _calculate_winners(self) -> None:
        logging.info(f"action: sorteo | result: in_progress")
        bets: list[Bet] = load_bets()

        for bet in bets:
            if has_won(bet):
                self._winners_per_agency[str(bet.agency)] = (
                    self._winners_per_agency.get(str(bet.agency), []) + [bet]
                )

        logging.info(f"action: sorteo | result: success")

    def get_winners(self, agency: str) -> list[Bet]:
        self._winners_ready.wait()

        with self._lock:
            return self._winners_per_agency.get(agency, [])


class BetService(Service):
    coordinator = BetCoordinator()

    def handle_message(self, message_encoded: bytes) -> Optional[bytes]:
        response: Message | None = None

        if BetMessage.matches(message_encoded):
            response = self._handle_bet_message(message_encoded)
        elif BetBatchMessage.matches(message_encoded):
            response = self._handle_bet_batch_message(message_encoded)
        elif EndOfBetsMessage.matches(message_encoded):
            response = self._handle_end_of_bets_message(message_encoded)
        else:
            raise ValueError("Message type not implemented")

        return response.serialize() if response else None

    def _handle_bet_message(self, message_encoded: bytes) -> AckMessage:
        try:
            bet_message: BetMessage = BetMessage.deserialize(message_encoded)

            logging.info(
                f"action: apuesta_almacenada | result: in_progress | dni: {bet_message.document} | numero: {bet_message.number}"
            )
            bet = Bet(
                agency=bet_message.agency,
                first_name=bet_message.first_name,
                last_name=bet_message.last_name,
                document=bet_message.document,
                birthdate=bet_message.birthdate,
                number=bet_message.number,
            )
            self.coordinator.store_bets([bet])
            logging.info(
                f"action: apuesta_almacenada | result: success | dni: {bet_message.document} | numero: {bet_message.number}"
            )
            return AckMessage(success=True)
        except Exception as e:
            return AckMessage(success=False)

    def _handle_bet_batch_message(self, message_encoded: bytes) -> AckMessage:
        bet_batch_message: BetBatchMessage = BetBatchMessage.deserialize(
            message_encoded
        )

        bets: list[Bet] = []

        for bet_message in bet_batch_message.bet_messages:
            bet = Bet(
                agency=bet_message.agency,
                first_name=bet_message.first_name,
                last_name=bet_message.last_name,
                document=bet_message.document,
                birthdate=bet_message.birthdate,
                number=bet_message.number,
            )
            bets.append(bet)

        try:
            logging.info(
                f"action: apuesta_recibida | result: in_progress | cantidad: {len(bets)}"
            )
            self.coordinator.store_bets(bets)
            logging.info(
                f"action: apuesta_recibida | result: success | cantidad: {len(bets)}"
            )
            return AckMessage(success=True)
        except Exception as e:
            logging.error(
                f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}"
            )
            return AckMessage(success=False)

    def _handle_end_of_bets_message(self, message_encoded: bytes) -> None:
        end_of_bets_message: EndOfBetsMessage = EndOfBetsMessage.deserialize(
            message_encoded
        )
        self.coordinator.mark_end_of_bets(end_of_bets_message.agency)

        winners: list[Bet] = self.coordinator.get_winners(end_of_bets_message.agency)

        return WinnersNotificationMessage(
            count=len(winners),
            documents=[winner.document for winner in winners],
        )
