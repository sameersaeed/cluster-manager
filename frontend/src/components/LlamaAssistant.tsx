import React, { useState } from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faMeta } from '@fortawesome/free-brands-svg-icons';
import toastr from '../toastr.js'; 
import '../toastr.css'; 
  
interface LlamaAssistantProps {
    yamlType: string;
    setYamlConfig: React.Dispatch<React.SetStateAction<string>>;
}

const LlamaAssistant: React.FC<LlamaAssistantProps> = ({ yamlType, setYamlConfig  }) => {
  const [groqQuery, setGroqQuery] = useState("");

  const handleLlamaRequest = async () => {
    if (!groqQuery.trim()) {
      toastr.error("Your Llama query cannot be empty");
      return;
    }

    try {
      const response = await fetch("http://localhost:8080/api/groq", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          yamlType, 
          query: groqQuery,
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to fetch Groq API response");
      }

      const data = await response.json();
      console.log(data);
      const fullResponse = data.choices[0].message.content;

      // simulate AI model typing effect for better interactivity
      setYamlConfig("");
      let index = 0;

      const typeEffect = () => {
        if (index < fullResponse.length) {
            const responseChar = fullResponse[index];
            if (responseChar !== 'undefined') { 
              setYamlConfig((prev) => prev + responseChar);
            }
            index++;
            setTimeout(typeEffect, 10);
          }
      };

      toastr.success("Request sent to Llama successfully");
      typeEffect();
    } catch (error) {
      console.error("Error:", error);
      toastr.error("Failed to generate YAML using Llama");
    }
  };

  return (
    <div className="mt-4 flex flex-col items-center">
      <input
        type="text"
        value={groqQuery}
        onChange={(e) => setGroqQuery(e.target.value)}
        placeholder={`Ask Llama AI to generate a ${yamlType} YAML`}
        className="w-3/4 p-2 border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 text-gray-900 dark:text-gray-200 rounded-md focus:ring focus:ring-blue-500"
      />
      <button
        onClick={handleLlamaRequest}
        className={`mt-3 py-2 px-4 rounded-md transition-colors ${
          groqQuery.length === 0
            ? "bg-gray-400 text-white cursor-not-allowed"
            : "bg-yellow-500 text-white hover:bg-yellow-600"
        }`}
        disabled={groqQuery.length === 0}
      >
        Ask Llama &nbsp;
        <FontAwesomeIcon icon={faMeta} className="text-white text-lg" />
      </button>
    </div>
  );
};

export default LlamaAssistant;
