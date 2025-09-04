from abc import ABC, abstractmethod

########################################################
# ABSTRACT MESSAGE
########################################################


class Message(ABC):
    _FIELD_DELIM = "^"
    _RECORD_DELIM = "~"

    @classmethod
    def get_tag(cls) -> str:
        return cls.__name__

    @classmethod
    @abstractmethod
    def matches(cls, message_encoded: bytes) -> bool: ...

    @classmethod
    @abstractmethod
    def deserialize(cls, message_encoded: bytes) -> "Message": ...

    @abstractmethod
    def serialize(self) -> bytes: ...


########################################################
# MESSAGES RECEIVED
########################################################


class BetMessage(Message):
    """
    obj = BetMessage("A1", "Sofia", "Sosa", "33828373", "1936-10-26", "8887")
    serialized = b"BetMessage^A1^Sofia^Sosa^33828373^1936-10-26^8887"
    """

    def __init__(
        self,
        agency: str,
        first_name: str,
        last_name: str,
        document: str,
        birthdate: str,
        number: str,
    ):
        self.agency: str = agency
        self.first_name: str = first_name
        self.last_name: str = last_name
        self.document: str = document
        self.birthdate: str = birthdate
        self.number: str = number

    @classmethod
    def matches(cls, message_encoded: bytes) -> bool:
        string = message_encoded.decode("utf-8")
        parts = string.split(cls._FIELD_DELIM)
        return parts and parts[0] == cls.get_tag()

    @classmethod
    def deserialize(cls, message_encoded: bytes) -> "BetMessage":
        string = message_encoded.decode("utf-8")
        fields: list = string.split(cls._FIELD_DELIM)

        if len(fields) != 7 or fields[0] != cls.get_tag():
            raise ValueError(f"Invalid {cls.get_tag()} format: {string!r}")

        return cls(*fields[1:])

    def serialize(self) -> bytes:
        raise NotImplementedError


class BetBatchMessage(Message):
    """
    obj = BetBatchMessage([
        BetMessage("A1", "Sofia", "Sosa", "33828373", "1936-10-26", "8887"),
        BetMessage("A1", "Agustin", "Perez", "12946214", "1934-08-15", "9495"),
    ])
    serialized = (
        b"BetBatchMessage"
        b"~BetMessage^A1^Sofia^Sosa^33828373^1936-10-26^8887"
        b"~BetMessage^A1^Agustin^Perez^12946214^1934-08-15^9495"
    )
    """

    def __init__(self, bet_messages: list[BetMessage]):
        self.bet_messages: list[BetMessage] = bet_messages

    @classmethod
    def matches(cls, message_encoded: bytes) -> bool:
        string = message_encoded.decode("utf-8")
        parts = string.split(cls._RECORD_DELIM)
        return parts and parts[0] == cls.get_tag()

    @classmethod
    def deserialize(cls, message_encoded: bytes) -> "BetBatchMessage":
        string = message_encoded.decode("utf-8")
        records: list[str] = string.split(cls._RECORD_DELIM)

        if len(records) < 2 or records[0] != cls.get_tag():
            raise ValueError(f"Invalid {cls.get_tag()} format: {string!r}")

        bet_messages: list[BetMessage] = [
            BetMessage.deserialize(record_str.encode("utf-8"))
            for record_str in records[1:]
        ]
        return cls(bet_messages=bet_messages)

    def serialize(self) -> bytes:
        raise NotImplementedError


class EndOfBetsMessage(Message):
    """
    obj = EndOfBetsMessage("A1")
    serialized = b"EndOfBetsMessage^A1"
    """

    def __init__(self, agency: str):
        self.agency: str = agency

    @classmethod
    def matches(cls, message_encoded: bytes) -> bool:
        s = message_encoded.decode("utf-8")
        parts = s.split(cls._FIELD_DELIM)
        return len(parts) == 2 and parts[0] == cls.get_tag()

    @classmethod
    def deserialize(cls, message_encoded: bytes) -> "EndOfBetsMessage":
        string = message_encoded.decode("utf-8")
        parts = string.split(cls._FIELD_DELIM)
        if len(parts) != 2 or parts[0] != cls.get_tag():
            raise ValueError(f"Invalid {cls.get_tag()} format: {string!r}")
        return cls(parts[1])

    def serialize(self) -> bytes:
        raise NotImplementedError


########################################################
# MESSAGES SENT
########################################################


class AckMessage(Message):
    """
    obj = AckMessage(True)
    serialized = b"AckMessage^True"
    """

    def __init__(self, success: bool):
        super().__init__()
        self.success: bool = success

    @classmethod
    def matches(cls, message_encoded: bytes) -> bool:
        raise NotImplementedError

    @classmethod
    def deserialize(cls, message_encoded: bytes) -> "AckMessage":
        raise NotImplementedError

    def serialize(self) -> bytes:
        string = self._FIELD_DELIM.join(
            [
                self.get_tag(),
                str(self.success),
            ]
        )
        return string.encode("utf-8")


class WinnersNotificationMessage(Message):
    """
    obj = WinnersNotificationMessage(3, [30904465, 21689196, 34407251])
    serialized = b"WinnersNotificationMessage^3^30904465~21689196~34407251"
    """

    def __init__(self, count: int, documents: list[int]):
        self.count: int = int(count)
        self.documents: list[int] = [int(d) for d in documents]

    @classmethod
    def matches(cls, message_encoded: bytes) -> bool:
        raise NotImplementedError

    @classmethod
    def deserialize(cls, message_encoded: bytes) -> "WinnersNotificationMessage":
        raise NotImplementedError

    def serialize(self) -> bytes:
        count_str = str(len(self.documents))
        docs_str = self._RECORD_DELIM.join(str(d) for d in self.documents)
        string = self._FIELD_DELIM.join([self.get_tag(), count_str, docs_str])
        return string.encode("utf-8")
