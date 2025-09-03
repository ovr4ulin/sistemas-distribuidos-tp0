from common.messages import BetMessage, AckMessage, Message
from common.utils import Bet, store_bets
from abc import ABC, abstractmethod
from typing import Optional
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
    def handle_message(self, message_encoded: bytes) -> Optional[bytes]:
        response: Message | None = None

        if BetMessage.matches(message_encoded):
            response = self._handle_bet_message(message_encoded)
        else:
            raise ValueError("Message type not implemented")

        return response.to_bytes() if response else None

    def _handle_bet_message(self, message_encoded: bytes) -> AckMessage:
        bet_message: BetMessage = BetMessage.from_bytes(message_encoded)

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

        return AckMessage(processed_count=1)
