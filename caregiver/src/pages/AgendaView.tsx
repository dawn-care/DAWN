import { useParams } from 'react-router-dom';
import { useState } from 'react';
import AgendaList, { type ConfigurationItem, type Module }  from '../components/AgendaList';

export default function AgendaView() {
    const { date } = useParams();
    const [items, setItems] = useState<ConfigurationItem[]>([
        { id: 1, carepartner_id: 1, module_id: 1, occurence_type: "daily", occurence_definition: ["monday", "tuesday"], position: 1 },
        { id: 2, carepartner_id: 1, module_id: 2, occurence_type: "daily", occurence_definition: ["monday", "tuesday"], position: 7 },
        { id: 3, carepartner_id: 1, module_id: 3, occurence_type: "daily", occurence_definition: ["monday", "tuesday"], position: 12 },
        { id: 4, carepartner_id: 1, module_id: 4, occurence_type: "daily", occurence_definition: ["monday", "tuesday"], position: 15 },
        { id: 5, carepartner_id: 1, module_id: 5, occurence_type: "daily", occurence_definition: ["monday", "tuesday"], position: 27 },
    ]);

    const [modules, setModules] = useState<Module[]>([
        {
            id: 1,
            module_type: 'media',
            carepartner_id: 1,
            metadata: {
                type: 'video',
                id: 1,
                uri: '',
                filename: 'Some Video',
                extension: 'mp4'
            }
        },
        {
            id: 2,
            module_type: 'media',
            carepartner_id: 1,
            metadata: {
                type: 'photo',
                id: 2,
                uri: '',
                filename: 'Vacation 2023',
                extension: 'jpg'
            }
        },
        {
            id: 3,
            module_type: 'widget',
            carepartner_id: 1,
            metadata: {
                type: 'weather',
                id: 3,
                zipcode: '90210'
            }
        },
        {
            id: 4,
            module_type: 'widget',
            carepartner_id: 1,
            metadata: {
                type: 'news',
                id: 4,
                category: 'Sports',
                language: 'en',
                country: 'US'
            }
        },
        {
            id: 5,
            module_type: 'prompt',
            carepartner_id: 1,
            metadata: {
                type: 'message',
                id: 5,
                message: 'Hope you are having a great day!',
                image_uri: 'https://example.com/greeting.jpg'
            }
        }
    ]);

    return (
        <div className="min-h-screen bg-white">
            <div className="px-8 py-8">
                <div className="flex items-center justify-between mb-8">
                    <button className="text-4xl text-gray-800 hover:text-gray-600">&lt;</button>
                    <h2 className="text-4xl font-light text-gray-800">{date} Videos for Dad</h2>
                    <button className="text-4xl text-gray-800 hover:text-gray-600">&gt;</button>
                </div>

                <div className="flex gap-8 mb-8">
                    <div className="flex items-center gap-4">
                        <label className="text-gray-700 font-medium">View:</label>
                        <button className="px-6 py-2 bg-gray-200 border-b-2 border-gray-400 rounded text-gray-800 flex items-center gap-2 hover:bg-gray-300">
                            Daily
                            <span>∨</span>
                        </button>
                    </div>
                    <div className="flex items-center gap-4">
                        <label className="text-gray-700 font-medium">Care Partner:</label>
                        <button className="px-6 py-2 bg-gray-200 border-b-2 border-gray-400 rounded text-gray-800 flex items-center gap-2 hover:bg-gray-300">
                            Dad
                            <span>∨</span>
                        </button>
                    </div>
                </div>

                <AgendaList items={items} setItems={setItems} modules={modules} setModules={setModules} />

                <div className="flex items-center justify-between mt-12">
                    <button className="text-3xl text-gray-600 hover:text-gray-800">⚙</button>
                    <div className="flex gap-6">
                        <button className="px-8 py-3 border-2 border-cyan-400 text-cyan-400 rounded-lg font-semibold hover:bg-cyan-50">
                            Preview
                        </button>
                        <button className="px-8 py-3 border-2 border-cyan-400 bg-white text-cyan-400 rounded-lg text-3xl font-bold hover:bg-cyan-50">
                            +
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}