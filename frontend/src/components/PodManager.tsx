import React, { useState, useEffect } from 'react';
import axios from 'axios';
import yaml from 'js-yaml';
import Modal from 'react-modal';
import LlamaAssistant from './LlamaAssistant';
import toastr from '../toastr.js'; 
import '../toastr.css'; 

interface Pod {
    name: string;
    status: string;
}

const PodManager: React.FC<{ namespace: string }> = ({ namespace }) => {
    const [pods, setPods] = useState<Pod[]>([]);
    const [modalPodName, setModalPodName] = useState<string>('');
    const [inputPodName, setInputPodName] = useState<string>('');
    const [podLogs, setPodLogs] = useState<string>('');
    const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
    const [modalType, setModalType] = useState<'logs' | 'create' | 'edit' | null>(null);
    const [yamlConfig, setYamlConfig] = useState<string>('');
    const apiUrl = process.env.REACT_APP_API_URL;

    // poll for updates in pod list every 10 seconds
    useEffect(() => {
        let pollingInterval: NodeJS.Timeout;

        const fetchPods = async () => {
            try {
                const response = await axios.get(`${apiUrl}/api/pods/${namespace}`);
                const updatedPodsList = response.data.pods || [];

                if (JSON.stringify(updatedPodsList) !== JSON.stringify(pods)) {
                    setPods(updatedPodsList);
                }
            } catch (error) {
                console.error("Error fetching pods:", error);
            }
        };

        if (namespace) {
            fetchPods();
            pollingInterval = setInterval(fetchPods, 10000);
        }

        return () => clearInterval(pollingInterval);
    }, [apiUrl, namespace, pods]);

    // different modal information is displayed based on the button pressed
    const openModal = (type: 'logs' | 'create' | 'edit', podName: string) => {
        setModalPodName(podName);
        setModalType(type);
        setIsModalOpen(true);
        if (type === 'logs') {
            handleGetPodLogs(podName);
        } else if (type === 'edit') {
            handleGetPodYaml(podName);
        }
    };

    const closeModal = () => {
        setIsModalOpen(false);
        setModalType(null);
    };

    // default yaml file for pod creation
    // the edit modal currently uses this default yaml as well, 
    // rather than the actual yaml being used by the deployed pod
    useEffect(() => {
        if (modalType !== 'logs') {
            setYamlConfig(
                `apiVersion: v1
kind: Pod
metadata:
  name: ${modalType === 'create' ? inputPodName : modalPodName}
spec:
  containers:
    - name: ${modalType === 'create' ? inputPodName : modalPodName}
      image: nginx`
            );
        }
    }, [inputPodName, modalPodName, modalType]);

    // sends a POST request to the backend to try to create a new pod
    const handleCreatePod = () => {
        try {
            yaml.load(yamlConfig);

            axios.post(`${apiUrl}/api/pod/${namespace}/${inputPodName}`, yamlConfig, {
                headers: {
                    'Content-Type': 'application/x-yaml'
                }})
                .then(response => {
                    
                    toastr.success(`A new pod called '${inputPodName}' has successfully been created in namespace '${namespace}'!`);
                    setPods([...pods, { name: inputPodName, status: 'Pending' }]);
                    closeModal();
                })
                .catch(error => {
                    console.error(`Error creating pod '${inputPodName}' in namespace '${namespace}':`, error);
                    toastr.error(`Failed to create pod '${inputPodName}': ${error}`);
                });
        } catch (error) {
            console.error('Error parsing YAML:', error);
            toastr.error('Invalid YAML format. Please correct it and try again.');
        }
    };

    /** 
     * sends a PUT request to the backend to try to create a new pod
     * as a workaround for certain pod properties being immutable, 
     * the backend will delete the pod and recreate it using the updated yaml
    */
    const handleUpdatePod = () => {
        try {
            yaml.load(yamlConfig);
            axios.put(`${apiUrl}/api/pod/${namespace}/${modalPodName}`, yamlConfig, {
                headers: {
                    'Content-Type': 'application/x-yaml',
                }
            })
                .then(response => {
                    toastr.success(`Pod '${modalPodName}' has been updated successfully in namespace '${namespace}'!`);
                    axios.get(`${apiUrl}/api/pods/${namespace}`)
                        .then(response => setPods(response.data.pods || []))
                        .catch(error => console.error(`Error fetching updated pod list for namespace ${namespace}:`, error));
                    closeModal();
                })
                .catch(error => {
                    console.error(`Error updating pod '${modalPodName}' in namespace '${namespace}':`, error);
                    toastr.error(`Failed to update pod '${inputPodName}': ${error}`);
                });
        } catch (error) {
            console.error('Error parsing YAML:', error);
            toastr.error('Invalid YAML format. Please correct it and try again.');
        }
    };

    // sends a DELETE request to the backend to try to delete the selected pod
    const handleDeletePod = (podName: string) => {
        axios.delete(`${apiUrl}/api/pod/${namespace}/${podName}`)
            .then(response => {
                toastr.success(`Pod '${podName}' has successfully been deleted in namespace '${namespace}'!`);
                setPods(pods.filter(pod => pod.name !== podName));
            })
            .catch(error => console.error(`Error deleting pod '${podName}':`, error));
    };

    // retrieves the selected pod's logs by sending a GET request to the backend
    const handleGetPodLogs = (podName: string) => {
        axios.get(`${apiUrl}/api/pod/${namespace}/${podName}/logs`)
            .then(response => {
                setPodLogs(response.data.logs)
            })
            .catch(error => console.error(`Error fetching pod logs for pod '${podName}' in namespace '${namespace}':`, error));
    };

    // this is currently unused as kubernetes adds extra properties to pod yamls 
    // as a result, pressing the edit button will just show a basic default yaml for now
    // with the pod name filled in, not the full yaml with all the additional properties
    const handleGetPodYaml = (podName: string) => {
        axios.get(`${apiUrl}/api/pod/${namespace}/${podName}/yaml`)
            .then(response => {
                if (response.data.yaml !== undefined) setYamlConfig(response.data);
            })
            .catch(error => console.error(`Error fetching YAML for pod '${namespace}':`, error));
    };

    return (
        <div className="bg-white dark:bg-gray-800 shadow-lg rounded-lg p-6 hover:scale-105 transition-transform">
            <h3 className="text-lg font-semibold mb-4">Manage Pods</h3>
            <input
                type="text"
                value={inputPodName}
                onChange={(e) => setInputPodName(e.target.value)}
                placeholder="Pod name"
                className="w-full mb-3 p-2 border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 text-gray-900 dark:text-gray-200 rounded-md focus:ring focus:ring-blue-500"
            />
            <button
                onClick={() => openModal('create', '')}
                className={`py-2 px-4 rounded-md transition-colors ${inputPodName.length === 0
                        ? 'bg-gray-400 text-white cursor-not-allowed'
                        : 'bg-blue-600 text-white hover:bg-blue-700'
                    }`}
                disabled={inputPodName.length === 0}
            >
                Create new pod
            </button>

            <h3 className="text-lg font-semibold mt-6 mb-2">Existing Pods</h3>
            <ul className="space-y-2">
                {pods.length === 0 ? (
                    <li className="flex items-center justify-between bg-gray-100 dark:bg-gray-700 p-3 rounded-md shadow">
                        No pods found.
                    </li>
                ) : (
                    pods.map((pod, index) => (
                        <li
                            key={index}
                            className="flex items-center justify-between bg-gray-100 dark:bg-gray-700 p-3 rounded-md shadow"
                        >
                            <span className="text-sm font-medium text-gray-800 dark:text-gray-200 truncated-name flex-1">
                                {pod.name}
                                <span
                                    className={`inline-block w-3 h-3 rounded-full ml-2 cursor-pointer ${
                                        pod.status === "Running"
                                            ? "bg-green-500"
                                            : pod.status === "Pending"
                                            ? "bg-gray-500"
                                            : "bg-red-500"
                                    }`}
                                    title={pod.status}  
                                ></span>
                            </span>
                            <div className="flex space-x-2 flex-shrink-0">
                                <button
                                    onClick={() => openModal('logs', pod.name)}
                                    className="bg-gray-500 text-white py-1 px-3 rounded-md hover:bg-gray-600 transition-colors"
                                >
                                    Logs
                                </button>
                                <button
                                    onClick={() => openModal('edit', pod.name)}
                                    className="bg-blue-500 text-white py-1 px-3 rounded-md hover:bg-blue-600 transition-colors"
                                >
                                    Edit
                                </button>
                                <button
                                    onClick={() => handleDeletePod(pod.name)}
                                    className="bg-red-500 text-white py-1 px-3 rounded-md hover:bg-red-600 transition-colors"
                                >
                                    Delete
                                </button>
                            </div>
                        </li>
                    ))
                )}
            </ul>

            <Modal
                isOpen={isModalOpen}
                onRequestClose={closeModal}
                contentLabel="Pod Modal"
                className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-xl w-1/2 fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2"
                overlayClassName="fixed inset-0 bg-gray-900 bg-opacity-50"
            >
                <h3 className="text-xl font-semibold mb-4 text-gray-900 dark:text-gray-100">
                    {modalType === 'logs' ? `Logs for ${modalPodName}` : modalType === 'create' ? 'Create pod' : `Edit pod '${modalPodName}'`}
                </h3>
                {modalType === 'logs' ? (
                    <pre className="bg-gray-50 dark:bg-gray-900 p-3 rounded-md max-h-96 overflow-y-auto text-gray-900 dark:text-gray-100">
                        {podLogs || 'Fetching logs...'}
                    </pre>
                ) : (
                    <>
                        <textarea
                            value={yamlConfig}
                            onChange={(e) => setYamlConfig(e.target.value)}
                            rows={25}
                            className="w-full p-3 bg-gray-50 dark:bg-gray-700 text-gray-900 dark:text-gray-200 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        />
                        <LlamaAssistant yamlType="pod" setYamlConfig={setYamlConfig} />
                    </>
                )}
                <div className="mt-4 flex justify-between">
                    <button
                        className="bg-gray-500 text-white py-2 px-4 rounded-md hover:bg-gray-600 transition-colors"
                        onClick={closeModal}
                    >
                        Cancel
                    </button>
                    {modalType !== 'logs' && (
                        <button
                            className="bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors"
                            onClick={modalType === 'create' ? handleCreatePod : modalType === 'edit' ? handleUpdatePod : undefined}
                        >
                            {modalType === 'create' ? 'Create' : 'Update'}
                        </button>
                    )}
                </div>
            </Modal>
        </div>
    );
};

export default PodManager;