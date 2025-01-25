import React, { useState, useEffect } from 'react';
import axios from 'axios';
import Modal from 'react-modal';
import yaml from 'js-yaml';

const DeploymentManager: React.FC<{ namespace: string }> = ({ namespace }) => {
  const [deployments, setDeployments] = useState<string[]>([]);
  const [deploymentName, setDeploymentName] = useState<string>('');
  const [yamlConfig, setYamlConfig] = useState<string>('');
  const [logs, setLogs] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);

  useEffect(() => {
    if (namespace) {
      axios.get(`http://localhost:8080/api/deployments/${namespace}`)
        .then(response => {
          setDeployments(response.data.deployments || []);
        })
        .catch(error => console.error(`Error fetching deployments from "${namespace}":`, error));
    }
  }, [namespace]);

  const openModal = () => {
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
  };

  useEffect(() => {
    setYamlConfig(
      `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${deploymentName}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${deploymentName}
  template:
    metadata:
      labels:
        app: ${deploymentName}
    spec:
      containers:
      - name: ${deploymentName}
        image: 'nginx'`);
  }, [deploymentName]);

  const handleCreateDeployment = () => {
    try {
      yaml.load(yamlConfig);

      axios.post(`http://localhost:8080/api/deployment/${namespace}/${deploymentName}`, yamlConfig, {
        headers: { 'Content-Type': 'application/x-yaml' }
      })
        .then(response => {
          alert(`A new deployment called ${deploymentName} has successfully been created in namespace ${namespace}!`);
          setDeployments([...deployments, deploymentName]);
          closeModal();
        })
        .catch(error => {
          console.error(`Error creating deployment "${deploymentName}":`, error);
          alert(`Failed to create deployment: ${error.message}`);
        });
    } catch (error) {
      console.error('Error parsing YAML:', error);
      alert('Invalid YAML format. Please correct it and try again.');
    }
  };

  const handleDeleteDeployment = (deploymentName: string) => {
    axios.delete(`http://localhost:8080/api/deployment/${namespace}/${deploymentName}`)
      .then(response => {
        alert(`Deployment ${deploymentName} has successfully been deleted in namespace ${namespace}!`);
        setDeployments(deployments.filter(dep => dep !== deploymentName));
      })
      .catch(error => console.error(`Error deleting deployment "${deploymentName}":`, error));
  };

  return (
    <div>
      <h2>Manage Deployments</h2>
      <input
        type="text"
        value={deploymentName}
        onChange={(e) => setDeploymentName(e.target.value)}
        placeholder="Deployment name"
      />
      <button onClick={openModal}>Create new deployment</button>

      <h3>Existing deployments</h3>
      <ul>
        {deployments.length === 0 ? (
          <li>No deployments found.</li>
        ) : (
          deployments.map((deployment, index) => (
            <li key={index}>
              {deployment}
              <button onClick={() => handleDeleteDeployment(deployment)}>Delete</button>
            </li>
          ))
        )}
      </ul>

      <Modal isOpen={isModalOpen} onRequestClose={closeModal}>
        <h3>Edit Deployment YAML Configuration</h3>
        <textarea
          value={yamlConfig}
          onChange={(e) => setYamlConfig(e.target.value)}
          rows={10}
          cols={50}
        />
        <div>
          <button onClick={handleCreateDeployment}>Submit</button>
          <button onClick={closeModal}>Cancel</button>
        </div>
      </Modal>
    </div>
  );
};

export default DeploymentManager;