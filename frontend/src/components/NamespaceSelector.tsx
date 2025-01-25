import React, { useEffect, useState } from 'react';
import axios from 'axios';

interface NamespaceSelectorProps {
  selectedNamespace: string;
  onNamespaceChange: (namespace: string) => void;
}

const NamespaceSelector: React.FC<NamespaceSelectorProps> = ({ selectedNamespace, onNamespaceChange }) => {
  const [namespaces, setNamespaces] = useState<string[]>([]);

  useEffect(() => {
    axios.get('http://localhost:8080/api/namespaces')
      .then(response => {
        setNamespaces(response.data.namespaces);
      })
      .catch(error => console.error("Error fetching namespaces:", error));
  }, []);

  return (
    <div>
      <label htmlFor="namespace">Namespace:</label>
      <select 
        id="namespace"
        value={selectedNamespace}
        onChange={(e) => onNamespaceChange(e.target.value)} 
      >
        <option value="" defaultValue="">Select namespace</option>
        {namespaces.map(namespace => (
          <option key={namespace} value={namespace}>
            {namespace}
          </option>
        ))}
      </select>
    </div>
  );
};

export default NamespaceSelector;
