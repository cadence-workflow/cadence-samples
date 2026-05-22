import cadence
from datetime import datetime
from dataclasses import dataclass

tools_registry = cadence.Registry()

@dataclass
class Flight:
    from_city: str
    to_city: str
    departure_date: datetime
    price: float
    airline: str
    flight_number: str
    seat_number: str

@tools_registry.activity(name="book_flight")
async def book_flight(from_city: str, to_city: str, departure_date: datetime) -> Flight:
    """
    Book a Flight tool: a pure mock for demo purposes
    """
    return Flight(from_city=from_city, to_city=to_city, departure_date=departure_date, price=100, airline="United", flight_number="123456", seat_number="12A")

@dataclass
class UberTrip:
    from_address: str
    to_address: str
    passengers: int
    price: float
    driver_name: str
    driver_phone: str
    driver_car: str
    driver_car_plate: str
    driver_car_color: str

@tools_registry.activity(name="book_uber")
async def book_uber(from_address: str, to_address: str, passengers: int) -> UberTrip:
    """
    Book a Uber ride from start address to the destination address. default passengers is 1.
    """
    return UberTrip(from_address=from_address, to_address=to_address, passengers=passengers, price=100, driver_name="John Doe", driver_phone="1234567890", driver_car="Toyota", driver_car_plate="1234567890", driver_car_color="Red")
