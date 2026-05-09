from __future__ import annotations

import json
import mimetypes
import os
import sys
import uuid
from itertools import cycle
from pathlib import Path
from random import Random
from urllib.error import HTTPError, URLError
from urllib.parse import quote_plus
from urllib.request import Request, urlopen


API_BASE_URL = os.getenv("GADGETSCOUT_API_URL", "http://localhost:8001").rstrip("/")
API_PREFIX = os.getenv("GADGETSCOUT_API_PREFIX", "/api/v1")
COUNTRY_CODE = os.getenv("GADGETSCOUT_SEED_COUNTRY", "uk").lower()
RUN_ID = os.getenv("GADGETSCOUT_SEED_RUN_ID", uuid.uuid4().hex[:8])
IMAGES_DIR = Path(__file__).resolve().parent / "images"
RANDOM = Random(42)


PRODUCTS = [
    ("Apple iPhone 16 Pro Max", "Smartphones", "Apple", 99900, 119900),
    ("Google Pixel 9 Pro", "Smartphones", "Google", 74900, 99900),
    ("Samsung Galaxy S25 Ultra", "Smartphones", "Samsung", 94900, 124900),
    ("OnePlus 13", "Smartphones", "OnePlus", 69900, 89900),
    ("Sony WH-1000XM5 Wireless Headphones", "Headphones", "Sony", 25900, 34900),
    ("Bose QuietComfort Ultra Headphones", "Headphones", "Bose", 28900, 44900),
    ("Apple AirPods Pro 2", "Earbuds", "Apple", 17900, 22900),
    ("Samsung Galaxy Buds3 Pro", "Earbuds", "Samsung", 15900, 21900),
    ("Apple Watch Series 10", "Smartwatches", "Apple", 32900, 39900),
    ("Google Pixel Watch 3", "Smartwatches", "Google", 24900, 34900),
    ("Garmin Forerunner 265", "Fitness Trackers", "Garmin", 29900, 42900),
    ("Apple MacBook Air 13-inch M3", "Laptops", "Apple", 89900, 109900),
    ("Dell XPS 13", "Laptops", "Dell", 99900, 129900),
    ("Lenovo Legion Pro 7i", "Laptops", "Lenovo", 179900, 229900),
    ("iPad Air 11-inch M2", "Tablets", "Apple", 54900, 59900),
    ("Samsung Galaxy Tab S10 Plus", "Tablets", "Samsung", 74900, 94900),
    ("Kindle Paperwhite", "E-Readers", "Amazon", 10900, 15900),
    ("Amazon Echo Show 8", "Smart Home", "Amazon", 8999, 14999),
    ("Amazon Echo Dot", "Smart Home", "Amazon", 2999, 5499),
    ("Ring Battery Video Doorbell", "Video Doorbells", "Ring", 5999, 9999),
    ("TP-Link Deco XE75 Mesh WiFi", "Mesh WiFi", "TP-Link", 21900, 29900),
    ("Netgear Nighthawk WiFi 7 Router", "Routers", "Netgear", 39900, 54900),
    ("Anker Prime Power Bank", "Power Banks", "Anker", 8999, 12999),
    ("Anker 737 Charger", "Chargers", "Anker", 6999, 9999),
    ("Samsung T9 Portable SSD 2TB", "External SSDs", "Samsung", 14900, 22900),
    ("SanDisk Extreme microSD 1TB", "Memory Cards", "SanDisk", 8499, 12999),
    ("Logitech MX Keys S", "Keyboards", "Logitech", 8999, 11999),
    ("Logitech MX Master 3S", "Mice", "Logitech", 7499, 9999),
    ("GoPro HERO13 Black", "Action Cameras", "GoPro", 32900, 39900),
    ("DJI Mini 4 Pro", "Drones", "DJI", 68900, 85900),
]


ARTICLE_TITLES = [
    "How to spot a real storefront offer before checkout",
    "The best headphone deals to watch this month",
    "What makes a smartphone discount worth buying",
    "Laptop deal checks that save money in the long run",
    "Smart home products that usually drop below retail",
    "Why retailer history matters when comparing deals",
    "The quick guide to deal scores on LeadStorefront",
    "When to buy earbuds instead of full-size headphones",
    "How shipping costs change the real deal price",
    "The safest way to compare marketplace storefront offers",
    "What to check before buying refurbished tech",
    "The tablet features worth paying extra for",
    "Gaming laptop discounts that are actually useful",
    "How to compare smartwatch offers across countries",
    "The hidden costs in smart home starter kits",
    "Why product timing matters for seasonal tech deals",
    "How to read review counts on discounted products",
    "When coupon codes beat headline discounts",
    "How to choose between retailer and affiliate links",
    "The best categories for quick storefront conversions",
    "How Amazon deal pages can hide the real price",
    "Why accessories often produce the strongest savings",
    "How to compare battery life claims on sale items",
    "When last year's flagship is the better deal",
    "How to check warranty details before buying tech",
    "The fast checklist for travel storefront offers",
    "How to avoid overpaying for storage upgrades",
    "Why bundle deals need extra scrutiny",
    "The smart way to compare WiFi router offers",
    "How to decide if a coupon code is worth using",
]


ARTICLE_CATEGORIES = [
    "Buying Guides",
    "Deal Roundups",
    "Reviews",
    "Comparisons",
    "How To",
    "News",
]


def main() -> int:
    images = image_paths()
    product_options = get_json("/admin/products/options")
    article_options = get_json("/admin/articles/options")
    country_id = find_country_id(product_options["countries"], COUNTRY_CODE)
    product_category_ids = id_map(product_options["categories"])
    article_category_ids = id_map(article_options["categories"])

    created_products = []
    for product in build_products(country_id, product_category_ids):
        response = post_json("/admin/products/create", product)
        created_products.append(response["product"])
        print(f"created product: {response['product']['name']}")

    created_articles = []
    for article, image in zip(build_articles(article_category_ids, created_products), cycle(images)):
        response = post_json("/admin/articles/create", article)
        created_article = response["article"]
        upload_file(f"/admin/articles/{created_article['ID']}/main-image", image, "main_image")
        created_articles.append(created_article)
        print(f"created article: {created_article['title']}")

    print(f"Seeded {len(created_products)} products and {len(created_articles)} articles via {API_BASE_URL}{API_PREFIX}.")
    return 0


def build_products(country_id: int, category_ids: dict[str, int]) -> list[dict]:
    products = []
    for index, (name, category, brand, current_price, original_price) in enumerate(PRODUCTS, start=1):
        discount = round((1 - (current_price / original_price)) * 100)
        amazon_url = f"https://www.amazon.co.uk/s?k={quote_plus(name)}"
        seed_name = f"{name} {RUN_ID}"
        products.append(
            {
                "name": seed_name,
                "description": (
                    f"API-only local seed deal for {name}. This item uses Amazon as the retailer "
                    "and is intended for testing LeadStorefront product pages."
                ),
                "brand": brand,
                "model_number": f"LS-{RUN_ID}-{index:02d}",
                "image_url": amazon_image_url(index),
                "product_url": amazon_url,
                "affiliate_url": f"{amazon_url}&tag=leadstorefront-local",
                "retailer_name": "Amazon UK",
                "retailer_url": "https://www.amazon.co.uk",
                "source": "local_seed_api",
                "external_id": f"local-api-product-{RUN_ID}-{index:02d}",
                "currency": "GBP",
                "current_price_cents": current_price,
                "original_price_cents": original_price,
                "shipping_cost_cents": RANDOM.choice([0, 299, 499, None]),
                "discount_percent": discount,
                "coupon_code": "LOCAL10" if index % 5 == 0 else "",
                "deal_score": RANDOM.randint(70, 98),
                "rating": round(RANDOM.uniform(4.0, 4.9), 1),
                "review_count": RANDOM.randint(45, 8500),
                "is_available": True,
                "is_featured": index <= 8,
                "country_id": country_id,
                "category_id": category_ids[category],
            }
        )
    return products


def build_articles(category_ids: dict[str, int], products: list[dict]) -> list[dict]:
    category_cycle = cycle(ARTICLE_CATEGORIES)
    articles = []
    for index, title in enumerate(ARTICLE_TITLES, start=1):
        product = products[(index - 1) % len(products)]
        category = next(category_cycle)
        seed_title = f"{title} {RUN_ID}"
        body = "\n\n".join(
            [
                f"{title} is an API-only local seed article for testing LeadStorefront editorial pages.",
                (
                    f"It references {product['name']} so product-linked article views can be tested "
                    "with realistic deal content."
                ),
                (
                    "Use this article to verify cards, metadata, image upload handling, public article "
                    "detail routes, and admin article workflows."
                ),
            ]
        )
        articles.append(
            {
                "author": "LeadStorefront Editors",
                "title": seed_title,
                "subtitle": f"API-created mock article {index} for local LeadStorefront testing.",
                "body": body,
                "image_url": "",
                "meta_title": f"{title} | LeadStorefront",
                "meta_description": f"API-only local seed article about {title.lower()}.",
                "meta_keywords": "products,deals,local seed,technology",
                "canonical_url": "",
                "is_published": True,
                "article_category_id": category_ids[category],
                "product_id": product["ID"],
            }
        )
    return articles


def get_json(path: str) -> dict:
    url = api_url(path)
    request = Request(url, headers={"Accept": "application/json"}, method="GET")
    return read_json(request)


def post_json(path: str, payload: dict) -> dict:
    request = Request(
        api_url(path),
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json", "Accept": "application/json"},
        method="POST",
    )
    return read_json(request)


def upload_file(path: str, file_path: Path, field_name: str) -> None:
    boundary = f"----leadstorefront-local-seed-{uuid.uuid4().hex}"
    content_type = mimetypes.guess_type(file_path.name)[0] or "application/octet-stream"
    file_bytes = file_path.read_bytes()
    body = b"".join(
        [
            f"--{boundary}\r\n".encode(),
            f'Content-Disposition: form-data; name="{field_name}"; filename="{file_path.name}"\r\n'.encode(),
            f"Content-Type: {content_type}\r\n\r\n".encode(),
            file_bytes,
            b"\r\n",
            f"--{boundary}--\r\n".encode(),
        ]
    )
    request = Request(
        api_url(path),
        data=body,
        headers={"Content-Type": f"multipart/form-data; boundary={boundary}", "Accept": "application/json"},
        method="POST",
    )
    read_json(request)


def read_json(request: Request) -> dict:
    try:
        with urlopen(request, timeout=30) as response:
            response_body = response.read().decode("utf-8")
    except HTTPError as error:
        error_body = error.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"{request.get_method()} {request.full_url} failed with {error.code}: {error_body}") from error
    except URLError as error:
        raise RuntimeError(f"{request.get_method()} {request.full_url} failed: {error}") from error

    if response_body == "":
        return {}
    return json.loads(response_body)


def api_url(path: str) -> str:
    path = path if path.startswith("/") else f"/{path}"
    return f"{API_BASE_URL}{API_PREFIX}{path}"


def image_paths() -> list[Path]:
    paths = sorted(path for path in IMAGES_DIR.iterdir() if path.suffix.lower() in {".jpg", ".jpeg", ".png", ".webp"})
    if not paths:
        raise RuntimeError(f"No seed images found in {IMAGES_DIR}")
    return paths


def find_country_id(countries: list[dict], code: str) -> int:
    for country in countries:
        if country["code"].lower() == code:
            return country["ID"]
    raise RuntimeError(f"Country {code!r} was not returned by the API.")


def id_map(records: list[dict]) -> dict[str, int]:
    return {record["name"]: record["ID"] for record in records}


def amazon_image_url(index: int) -> str:
    # Product uploads are not supported yet, so use deterministic external placeholders.
    return f"https://m.media-amazon.com/images/I/71-local-seed-{index:02d}.jpg"


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(exc, file=sys.stderr)
        raise SystemExit(1)
