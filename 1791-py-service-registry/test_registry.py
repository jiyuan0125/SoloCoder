import unittest
import time
from unittest.mock import patch
import semantic_version

from main import ServiceRegistry


class TestServiceRegistry(unittest.TestCase):
    def setUp(self):
        self.registry = ServiceRegistry(expiration_seconds=30)

    def test_register(self):
        instance_id = self.registry.register(
            name="user-service",
            address="http://127.0.0.1:8080",
            metadata={"labels": {"env": "dev", "zone": "east"}},
            version="1.0.0"
        )
        self.assertIsNotNone(instance_id)
        self.assertEqual(len(self.registry.get_all_instances()), 1)

    def test_heartbeat_success(self):
        instance_id = self.registry.register(
            name="user-service",
            address="http://127.0.0.1:8080",
            metadata={}
        )
        self.assertTrue(self.registry.heartbeat(instance_id))

    def test_heartbeat_not_found(self):
        self.assertFalse(self.registry.heartbeat("non-existent-id"))

    def test_deregister(self):
        instance_id = self.registry.register(
            name="user-service",
            address="http://127.0.0.1:8080",
            metadata={}
        )
        self.assertEqual(len(self.registry.get_all_instances()), 1)
        self.assertTrue(self.registry.deregister(instance_id))
        self.assertEqual(len(self.registry.get_all_instances()), 0)

    def test_deregister_not_found(self):
        self.assertFalse(self.registry.deregister("non-existent-id"))

    def test_query_by_name(self):
        self.registry.register("service-a", "addr1", {})
        self.registry.register("service-b", "addr2", {})
        
        results = self.registry.query(name="service-a")
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0]["name"], "service-a")

    def test_query_by_labels(self):
        self.registry.register(
            "service-a", "addr1",
            {"labels": {"env": "prod", "zone": "east"}}
        )
        self.registry.register(
            "service-a", "addr2",
            {"labels": {"env": "dev", "zone": "east"}}
        )
        
        results = self.registry.query(labels={"env": "prod", "zone": "east"})
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0]["address"], "addr1")

    def test_query_by_version_range(self):
        self.registry.register("service-a", "addr1", {}, version="1.5.0")
        self.registry.register("service-a", "addr2", {}, version="2.1.0")
        self.registry.register("service-a", "addr3", {}, version="2.9.0")
        self.registry.register("service-a", "addr4", {}, version="3.0.0")
        
        results = self.registry.query(min_version="2.0.0", max_version="3.0.0")
        addresses = {r["address"] for r in results}
        self.assertIn("addr2", addresses)
        self.assertIn("addr3", addresses)
        self.assertNotIn("addr1", addresses)
        self.assertNotIn("addr4", addresses)

    def test_query_no_version_excludes_unversioned(self):
        self.registry.register("service-a", "addr1", {}, version="2.1.0")
        self.registry.register("service-a", "addr2", {}, version=None)
        
        results = self.registry.query(min_version="2.0.0")
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0]["address"], "addr1")

    def test_query_empty_result_returns_empty(self):
        results = self.registry.query(name="non-existent")
        self.assertEqual(len(results), 0)

    def test_is_expired(self):
        instance_id = self.registry.register("service-a", "addr", {})
        instance = self.registry.get_instance(instance_id)
        
        self.assertFalse(self.registry.is_expired(instance))
        
        with patch('time.time', return_value=time.time() + 40):
            self.assertTrue(self.registry.is_expired(instance))

    def test_cleanup_expired(self):
        self.registry.register("service-a", "addr1", {})
        
        with patch('time.time', return_value=time.time() + 40):
            cleaned = self.registry.cleanup_expired()
            self.assertEqual(cleaned, 1)
            self.assertEqual(len(self.registry.get_all_instances()), 0)


if __name__ == "__main__":
    unittest.main()
