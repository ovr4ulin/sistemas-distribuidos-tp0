from common.messages import (
    BetMessage,
    AckMessage,
    Message,
    BetBatchMessage,
    EndOfBetsMessage,
    WinnersRequestMessage,
    WinnersPendingMessage,
    WinnersNotificationMessage,
)
from common.utils import Bet, store_bets, load_bets, has_won
from abc import ABC, abstractmethod
from typing import Optional
import os
import logging

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


class BetService(Service):
    winners_per_agency: dict[str, int] = {}
    agencies_proccesed: set = set()
    total_agencies: int = int(os.environ.get("TOTAL_AGENCIES"))

    def handle_message(self, message_encoded: bytes) -> Optional[bytes]:
        response: Message | None = None

        if BetMessage.matches(message_encoded):
            response = self._handle_bet_message(message_encoded)
        elif BetBatchMessage.matches(message_encoded):
            response = self._handle_bet_batch_message(message_encoded)
        elif EndOfBetsMessage.matches(message_encoded):
            response = self._handle_end_of_bets_message(message_encoded)
        elif WinnersRequestMessage.matches(message_encoded):
            response = self._handle_winners_request_message(message_encoded)
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
            store_bets([bet])
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
            store_bets(bets)
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
        self.agencies_proccesed.add(end_of_bets_message.agency)

        if len(self.agencies_proccesed) < self.total_agencies:
            return None

        logging.info(f"action: sorteo | result: in_progress")

        bets: list[Bet] = load_bets()

        for bet in bets:
            if has_won(bet):
                self.winners_per_agency[str(bet.agency)] = self.winners_per_agency.get(
                    str(bet.agency), []
                ) + [bet]

        logging.info(
            f"action: sorteo | result: success | winners: {self.winners_per_agency}"
        )

    def _handle_winners_request_message(self, message_encoded: bytes) -> None:
        winners_request_message: WinnersRequestMessage = (
            WinnersRequestMessage.deserialize(message_encoded)
        )

        if len(self.agencies_proccesed) < self.total_agencies:
            return WinnersPendingMessage()

        return WinnersNotificationMessage(
            count=len(self.winners_per_agency.get(winners_request_message.agency, [])),
            documents=[
                winner.document
                for winner in self.winners_per_agency.get(
                    winners_request_message.agency, []
                )
            ],
        )
