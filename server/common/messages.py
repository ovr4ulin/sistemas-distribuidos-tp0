from abc import ABC
import inspect

########################################################
# ABSTRACT MESSAGE
########################################################


class Message(ABC):
    _DELIMITER: str = "|"
    _TAG: str = ""

    @classmethod
    def get_tag(cls) -> str:
        return cls._TAG or cls.__name__

    @classmethod
    def matches(cls, message_encoded: bytes) -> bool:
        string = message_encoded.decode("utf-8")
        parts = string.split(cls._DELIMITER)
        return parts and parts[0] == cls.get_tag()

    @classmethod
    def from_bytes(cls, message_encoded: bytes) -> "Message":
        string = message_encoded.decode("utf-8")
        fields: list = string.split(cls._DELIMITER)
        return cls(*fields[1:])

    def to_bytes(self) -> bytes:
        signature = inspect.signature(self.__class__.__init__)
        param_names = [name for name in signature.parameters if name != "self"]
        param_values = [getattr(self, param_name) for param_name in param_names]
        string = self._DELIMITER.join([self.get_tag()] + param_values)
        return string.encode("utf-8")


########################################################
# MESSAGES
########################################################


class BetMessage(Message):
    def __init__(
        self,
        agency: str,
        first_name: str,
        last_name: str,
        document: str,
        birthdate: str,
        number: str,
    ):
        super().__init__()
        self.agency: str = agency
        self.first_name: str = first_name
        self.last_name: str = last_name
        self.document: str = document
        self.birthdate: str = birthdate
        self.number: str = number


class AckMessage(Message):
    def __init__(self, processed_count: int):
        super().__init__()
        self.processed_count: str = str(processed_count)
