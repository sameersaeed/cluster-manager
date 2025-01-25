import React, { useState, useEffect } from 'react';
import axios from 'axios';
import Modal from 'react-modal';
import yaml from 'js-yaml';

const DeploymentManager: React.FC<{ namespace: string }> = ({ namespace }) => {
  const [deployments, setDeployments] = useState<string[]>([]);
  const [deploymentName, setDeploymentName] = useState<string>('');
  const [yamlConfig, setYamlConfig] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const apiUrl = process.env.REACT_APP_API_URL;

  useEffect(() => {
    if (namespace) {
      axios.get(`${apiUrl}/api/deployments/${namespace}`)
        .then(response => {
          setDeployments(response.data.deployments || []);
        })
        .catch(error => console.error(`Error fetching deployments from "${namespace}":`, error.response.data));
    }
  }, [apiUrl, namespace]);

  const openModal = () => {
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
  };

  // default yaml file for deployment creation
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
        image: nginx`);
  }, [deploymentName]);

  // sends a POST request to the backend to try to create a new deployment
  const handleCreateDeployment = () => {
    try {
      yaml.load(yamlConfig);

      axios.post(`${apiUrl}/api/deployment/${namespace}/${deploymentName}`, yamlConfig, {
        headers: { 'Content-Type': 'application/x-yaml' }
      })
        .then(response => {
          alert(`A new deployment called ${deploymentName} has successfully been created in namespace ${namespace}!`);
          setDeployments([...deployments, deploymentName]);
          closeModal();
        })
        .catch(error => {
          console.error(`Error creating deployment "${deploymentName}":`, error.response.data);
          alert(`Failed to create deployment: ${error.response.data}`);
        });
    } catch (error) {
      console.error('Error parsing YAML:', error);
      alert('Invalid YAML format. Please correct it and try again.');
    }
  };

  // sends a DELETE request to the backend to try to delete the selected deployment
  const handleDeleteDeployment = (deploymentName: string) => {
    axios.delete(`${apiUrl}/api/deployment/${namespace}/${deploymentName}`)
      .then(response => {
        alert(`Deployment ${deploymentName} has successfully been deleted in namespace ${namespace}!`);
        setDeployments(deployments.filter(dep => dep !== deploymentName));
      })
      .catch(error => console.error(`Error deleting deployment "${deploymentName}":`, error.response.data));
  };


  return (
    <div className="bg-white dark:bg-gray-800 shadow-lg rounded-lg p-6 hover:scale-105 transition-transform">
      <h3 className="text-lg font-semibold mb-4">Manage deployments</h3>
      <input
        type="text"
        value={deploymentName}
        onChange={(e) => setDeploymentName(e.target.value)}
        placeholder="Deployment name"
        className="w-full mb-3 p-2 border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 text-gray-900 dark:text-gray-200 rounded-md focus:ring focus:ring-blue-500"
      />
      <button
        onClick={openModal}
        className={`py-2 px-4 rounded-md transition-colors ${deploymentName.length === 0
            ? 'bg-gray-400 text-white cursor-not-allowed'
            : 'bg-blue-600 text-white hover:bg-blue-700'
          }`}
        disabled={deploymentName.length === 0}
      >
        Create new deployment
      </button>

      <h3 className="text-lg font-semibold mt-6 mb-3">Existing deployments</h3>
      <ul className="space-y-2">
        {deployments.length === 0 ? (
          <li className="flex items-center justify-between bg-gray-100 dark:bg-gray-700 p-3 rounded-md shadow">
            No deployments found.
          </li>
        ) : (
          deployments.map((deployment, index) => (
            <li
              key={index}
              className="flex items-center justify-between bg-gray-100 dark:bg-gray-700 p-3 rounded-md shadow"
            >
              <span className="text-sm font-medium text-gray-800 dark:text-gray-200 truncated-name">
                {deployment}
              </span>
              <button
                onClick={() => handleDeleteDeployment(deployment)}
                className="bg-red-600 text-white py-1 px-3 rounded-md hover:bg-red-700 transition-colors"
              >
                Delete
              </button>
            </li>
          ))
        )}
      </ul>

      <Modal
        isOpen={isModalOpen}
        onRequestClose={closeModal}
        className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-xl"
        overlayClassName="fixed inset-0 bg-gray-900 bg-opacity-50"
      >
        <h3 className="text-xl font-semibold mb-4 text-gray-900 dark:text-gray-100">
          Create new deployment
        </h3>
        <textarea
          value={yamlConfig}
          onChange={(e) => setYamlConfig(e.target.value)}
          rows={10}
          cols={50}
          className="w-full p-3 bg-gray-50 dark:bg-gray-700 text-gray-900 dark:text-gray-200 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
        />
        <div className="mt-4 flex justify-between">
          <button
            className="bg-gray-500 text-white py-2 px-4 rounded-md hover:bg-gray-600 transition-colors"
            onClick={closeModal}
          >
            Cancel
          </button>
          <button
            className="bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors"
            onClick={handleCreateDeployment}
          >
            Submit
          </button>
        </div>
      </Modal>
    </div>
  );
};

export default DeploymentManager;