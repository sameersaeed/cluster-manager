import React, { useState, useEffect } from 'react';
import axios from 'axios';
import Modal from 'react-modal';
import yaml from 'js-yaml';

interface Pod {
    name: string;
    status: string;
}

const PodManager: React.FC<{ namespace: string }> = ({ namespace }) => {
    const [pods, setPods] = useState<Pod[]>([]);
    const [podName, setPodName] = useState<string>('');
    const [logs, setLogs] = useState<string>('');
    const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
    const [yamlConfig, setYamlConfig] = useState<string>('');
    const [isEditing, setIsEditing] = useState<boolean>(false);

    useEffect(() => {
        if (namespace) {
            axios.get(`http://localhost:8080/api/pods/${namespace}`)
                .then(response => setPods(response.data.pods || []))
                .catch(error => console.error("Error fetching pods:", error));
        }
    }, [namespace]);

    const openModal = () => {
        setIsModalOpen(true);
        setIsEditing(false);
    };

    const closeModal = () => {
        setIsModalOpen(false);
    };

    useEffect(() => {
        setYamlConfig(
            `apiVersion: v1
kind: Pod
metadata:
  name: ${podName}
spec:
  containers:
    - name: ${podName}
      image: 'nginx'`);
    }, [podName]);

    const handleCreatePod = () => {
        try {
            yaml.load(yamlConfig);

            axios.post(`http://localhost:8080/api/pod/${namespace}/${podName}`, yamlConfig, {
                headers: {
                    'Content-Type': 'application/x-yaml'
                }
            })
                .then(response => {
                    alert(`A new pod called ${podName} has successfully been created in namespace ${namespace}!`);
                    setPods([...pods, { name: podName, status: 'Pending' }]);
                    closeModal();
                })
                .catch(error => {
                    console.error(`Error creating pod "${podName}":`, error);
                    alert(`Failed to create pod: ${error.message}`);
                });
        } catch (error) {
            console.error('Error parsing YAML:', error);
            alert('Invalid YAML format. Please correct it and try again.');
        }
    };

    const handleUpdatePod = () => {
        try {
            yaml.load(yamlConfig);
            axios.put(`http://localhost:8080/api/pod/${namespace}/${podName}`, yamlConfig, {
                headers: {
                    'Content-Type': 'application/x-yaml',
                }
            })
                .then(response => {
                    alert(`Pod ${podName} has been updated successfully!`);
                    axios.get(`http://localhost:8080/api/pod/${namespace}/${podName}/yaml`)
                        .then(() => {
                            setTimeout(() => {
                                axios.get(`http://localhost:8080/api/pods/${namespace}`)
                                .then(response => setPods(response.data.pods))
                                closeModal();
                            }, 2000);
                        })
                })
                .catch(error => {
                    console.error(`Error updating pod "${podName}":`, error);
                    alert(`Failed to update pod: ${error.message}`);
                });
        } catch (error) {
            console.error('Error parsing YAML:', error);
            alert('Invalid YAML format. Please correct it and try again.');
        }
    };

    const handleDeletePod = (podName: string) => {
        axios.delete(`http://localhost:8080/api/pod/${namespace}/${podName}`)
            .then(response => {
                alert(`Pod ${podName} has successfully been deleted in namespace ${namespace}!`);
                setPods(pods.filter(pod => pod.name !== podName));
            })
            .catch(error => console.error(`Error deleting pod "${podName}":`, error));
    };

    const handleGetPodLogs = (podName: string) => {
        axios.get(`http://localhost:8080/api/pod/${namespace}/${podName}/logs`)
            .then(response => setLogs(response.data.logs))
            .catch(error => console.error(`Error fetching pod logs for pod ${podName} in namespace ${namespace}:`, error));
    };

    const handleGetPodYaml = (podName: string) => {
        axios.get(`http://localhost:8080/api/pod/${namespace}/${podName}/yaml`)
            .then(response => {
                setYamlConfig(response.data.yaml || '');
                setPodName(podName);
                setIsModalOpen(true);
                setIsEditing(true);
            })
            .catch(error => console.error("Error fetching pod YAML:", error));
    };

    return (
        <div>
            <h2>Manage Pods</h2>
            <input
                type="text"
                value={podName}
                onChange={(e) => setPodName(e.target.value)}
                placeholder="Pod name"
            />
            <button onClick={openModal}>Create new pod</button>

            <h3>Pods</h3>
            <ul>
                {pods.length === 0 ? (
                    <li>No pods found.</li>
                ) : (
                    pods.map((pod) => (
                        <li key={pod.name}>
                            {pod.name} - Status: {pod.status}
                            <button onClick={() => handleGetPodYaml(pod.name)}>Edit</button>
                            <button onClick={() => handleDeletePod(pod.name)}>Delete</button>
                            <button onClick={() => handleGetPodLogs(pod.name)}>Pod logs</button>
                        </li>
                    ))
                )}
            </ul>

            <Modal isOpen={isModalOpen} onRequestClose={closeModal}>
                <h3>{isEditing ? `Update YAML for ${podName}` : 'Create new pod'}</h3>
                {isEditing && <h4>Note: Editing a pod will cause it to be deleted and re-created</h4>}
                <textarea
                    value={yamlConfig}
                    onChange={(e) => setYamlConfig(e.target.value)}
                    rows={10}
                    cols={50}
                />
                <div>
                    <button onClick={isEditing ? handleUpdatePod : handleCreatePod}>
                        {isEditing ? 'Update pod' : 'Create pod'}
                    </button>
                    <button onClick={closeModal}>Cancel</button>
                </div>
            </Modal>

            {logs && <div><h4>Pod logs ({podName})</h4><pre>{logs}</pre></div>}
        </div>
    );
};

export default PodManager;