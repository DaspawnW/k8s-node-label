package controller

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

var MasterNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-master-node",
	},
	Spec: v1.NodeSpec{
		Taints: []v1.Taint{
			{
				Key:    NodeRoleMasterLabel,
				Effect: "NoSchedule",
			},
		},
	},
}
var SoptMasterNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-spot-master",
	},
	Spec: v1.NodeSpec{
		ProviderID: "aws:///eu-central-1/i-123uzu123",
		Taints: []v1.Taint{
			{
				Key:    NodeRoleMasterLabel,
				Effect: "NoSchedule",
			},
		},
	},
}
var WorkerNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-worker-node",
	},
	Spec: v1.NodeSpec{
		ProviderID: "aws:///eu-central-1/i-123qwe123",
	},
}
var SpotWorkerNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-spot-node",
	},
	Spec: v1.NodeSpec{
		ProviderID: "aws:///eu-central-1/i-123asd132",
	},
}
var UnManagedNode = &v1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-unmanaged-node",
	},
	Spec: v1.NodeSpec{},
}

func TestHandlerShouldSetNodeRoleMaster(t *testing.T) {
	clientset := fake.NewSimpleClientset(MasterNode)
	testingMockDiscovery := TestingMockDiscovery{}

	c := NewNodeController(clientset, testingMockDiscovery, false, false, false)
	c.handler(MasterNode)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-master-node", metav1.GetOptions{})
	if _, ok := foundNode.Labels[NodeRoleMasterLabel]; !ok {
		t.Errorf("Expected label %s on node %s, but was not assigned", NodeRoleMasterLabel, "test-master-node")
	}
}

func TestHandlerShouldSetSpotMasterRole(t *testing.T) {
	clientset := fake.NewSimpleClientset(SoptMasterNode)
	testingMockDiscovery := TestingMockDiscovery{}

	c := NewNodeController(clientset, testingMockDiscovery, false, false, false)
	c.handler(SoptMasterNode)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-spot-master", metav1.GetOptions{})
	if _, ok := foundNode.Labels[NodeRoleSpotMasterLabel]; !ok {
		t.Errorf("Expected label %s on node %s, but was not assigned", NodeRoleSpotMasterLabel, "test-master-node")
	}
}

func TestHandlerShouldSetWorkerRoleIfWorker(t *testing.T) {
	clientset := fake.NewSimpleClientset(WorkerNode)
	testingMockDiscovery := TestingMockDiscovery{}

	c := NewNodeController(clientset, testingMockDiscovery, false, false, false)
	c.handler(WorkerNode)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-worker-node", metav1.GetOptions{})
	if _, ok := foundNode.Labels[NodeRoleMasterLabel]; ok {
		t.Errorf("Expected no label %s on node %s, but was assigned", NodeRoleMasterLabel, "test-worker-node")
	}
	if _, ok := foundNode.Labels[NodeRoleWorkerLabel]; !ok {
		t.Errorf("Expected label %s on node %s, but was not assigned", NodeRoleWorkerLabel, "test-worker-node")
	}
}

func TestHandlerShouldSetSpotWorkerRoleIfSpotWorker(t *testing.T) {
	clientset := fake.NewSimpleClientset(SpotWorkerNode)
	testingMockDiscovery := TestingMockDiscovery{}

	c := NewNodeController(clientset, testingMockDiscovery, false, false, false)
	c.handler(SpotWorkerNode)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-spot-node", metav1.GetOptions{})
	if _, ok := foundNode.Labels[NodeRoleMasterLabel]; ok {
		t.Errorf("Expected no label %s on node %s, but was assigned", NodeRoleMasterLabel, "test-spot-node")
	}
	if _, ok := foundNode.Labels[NodeRoleWorkerLabel]; ok {
		t.Errorf("Expected no label %s on node %s, but was assigned", NodeRoleWorkerLabel, "test-spot-node")
	}
	if _, ok := foundNode.Labels[NodeRoleSpotWorkerLabel]; !ok {
		t.Errorf("Expected label %s on node %s, but was not assigned", NodeRoleSpotWorkerLabel, "test-spot-node")
	}
}

func TestHandlerShouldSetWorkerRoleIfNotSet(t *testing.T) {
	clientset := fake.NewSimpleClientset(UnManagedNode)
	testingMockDiscovery := TestingMockDiscovery{}

	c := NewNodeController(clientset, testingMockDiscovery, false, false, false)
	c.handler(UnManagedNode)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-unmanaged-node", metav1.GetOptions{})
	if _, ok := foundNode.Labels[NodeRoleWorkerLabel]; !ok {
		t.Errorf("Expected label %s on node %s, but was not assigned", NodeRoleWorkerLabel, "test-unmanaged-node")
	}
}

func TestHandlerShouldPreventMasterFromLoadbalancing(t *testing.T) {
	clientset := fake.NewSimpleClientset(MasterNode)
	testingMockDiscovery := TestingMockDiscovery{}

	c := NewNodeController(clientset, testingMockDiscovery, true, true, false)
	c.handler(MasterNode)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-master-node", metav1.GetOptions{})
	if val, ok := foundNode.Labels[ExcludeLoadBalancerLabel]; ok {
		if val != "true" {
			t.Errorf("Expected label %s value 'true', but value was %s", ExcludeLoadBalancerLabel, val)
		}
	} else {
		t.Errorf("Expected label %s on node, but was not assigned", ExcludeLoadBalancerLabel)
	}

	if val, ok := foundNode.Labels[AlphaExcludeLoadBalancerLabel]; ok {
		if val != "true" {
			t.Errorf("Expected label %s value 'true', but value was %s", AlphaExcludeLoadBalancerLabel, val)
		}
	} else {
		t.Errorf("Expected label %s on node, but was not assiged", AlphaExcludeLoadBalancerLabel)
	}
}

func TestHandlerShouldExcludeNodeFromEviction(t *testing.T) {
	clientset := fake.NewSimpleClientset(MasterNode)
	testingMockDiscovery := TestingMockDiscovery{}

	c := NewNodeController(clientset, testingMockDiscovery, false, false, true)
	c.handler(MasterNode)

	foundNode, _ := clientset.CoreV1().Nodes().Get(context.TODO(), "test-master-node", metav1.GetOptions{})
	if val, ok := foundNode.Labels[ExcludeDisruptionLabel]; ok {
		if val != "true" {
			t.Errorf("Expected label %s value 'true', but value was %s", ExcludeDisruptionLabel, val)
		}
	} else {
		t.Errorf("Expected label %s on node, but was not assigned", ExcludeDisruptionLabel)
	}
}

type TestingMockDiscovery struct{}

func (TestingMockDiscovery) IsSpotInstance(node *v1.Node) bool {
	if node.Spec.ProviderID != "" && (node.Spec.ProviderID == "aws:///eu-central-1/i-123uzu123" || node.Spec.ProviderID == "aws:///eu-central-1/i-123asd132") {
		return true
	}
	return false
}
