import React, { useEffect, useState } from 'react';
import axios from 'axios';

interface NamespaceSelectorProps {
  selectedNamespace: string;
  onNamespaceChange: (namespace: string) => void;
}

const NamespaceSelector: React.FC<NamespaceSelectorProps> = ({ selectedNamespace, onNamespaceChange }) => {
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const apiUrl = process.env.REACT_APP_API_URL;

  // sends a GET request to the backend to fetch the namespaces from the cluster
  useEffect(() => {
    axios.get(`${apiUrl}/api/namespaces`)
      .then((response) => {
        setNamespaces(response.data.namespaces);
      })
      .catch((error) => console.error('Error fetching namespaces:', error));
  }, [apiUrl]);

  return (
    <div>
      <label htmlFor="namespace" className="block text-sm font-medium mb-2">
        Namespace
      </label>
      <select
        id="namespace"
        className="w-full bg-gray-50 dark:bg-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 text-gray-900 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 transition-all hover:ring-2 hover:ring-blue-300"
        value={selectedNamespace}
        onChange={(e) => onNamespaceChange(e.target.value)}
      >
        <option value="" disabled>Select namespace</option>
        {namespaces.map((namespace) => (
          <option key={namespace} value={namespace}>{namespace}</option>
        ))}
      </select>
    </div>
  );
};

export default NamespaceSelector;
