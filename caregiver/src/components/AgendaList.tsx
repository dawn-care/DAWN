import React from 'react';
import videoCameraIcon from '../assets/module_symbols/video-camera-svgrepo-com.svg';
import widgetIcon from '../assets/module_symbols/widget-svgrepo-com.svg';
import promptIcon from '../assets/module_symbols/prompt-edit-svgrepo-com.svg';

export type ModuleMetadata =
    | { type: 'photo'; id: number; uri: string; filename: string; extension: string }
    | { type: 'video'; id: number; uri: string; filename: string; extension: string }
    | { type: 'audio'; id: number; uri: string; filename: string; extension: string }
    | { type: 'weather'; id: number; zipcode: string }
    | { type: 'news'; id: number; category: string; language: string; country: string }
    | { type: 'carepartner_location'; id: number; location: string; message: string; image_uri: string }
    | { type: 'caregiver_location'; id: number; location: string; message: string; image_uri: string }
    | { type: 'message'; id: number; message: string; image_uri: string }
    | { type: 'event'; id: number; list_id: number; time: string; text: string };

export type Module = {
    id: number;
    module_type: string;
    caregiver_id?: number | null;
    carepartner_id?: number | null;
    metadata: ModuleMetadata | null;
};

export type ConfigurationItem = {
    id?: number;
    carepartner_id: number;
    module_id: number;
    occurence_type: string;
    occurence_definition: string[];
    position: number;
}

interface AgendaListProps {
    items: ConfigurationItem[];
    setItems: (items: ConfigurationItem[]) => void;
    modules: Module[];
    setModules: (modules: Module[]) => void;
}

export default function AgendaList({ items, setItems, modules, setModules }: AgendaListProps) {
    const [draggedIndex, setDraggedIndex] = React.useState<number | null>(null);

    const getModuleTitle = (moduleId: number): string => {
        const module = modules.find(m => m.id === moduleId);
        if (!module?.metadata) return `Module ${moduleId}`;

        const metadata = module.metadata;
        switch (metadata.type) {
            case 'video':
                return metadata.filename;
            case 'photo':
                return metadata.filename;
            case 'audio':
                return metadata.filename;
            case 'weather':
                return `Weather - Zip Code: ${metadata.zipcode}`;
            case 'news':
                return `News - Category: ${metadata.category}`;
            case 'message':
                return 'Message';
            case 'event':
                return `Event - ${metadata.text}`;
            case 'carepartner_location':
                return `Location - ${metadata.location}`;
            case 'caregiver_location':
                return `Location - ${metadata.location}`;
            default:
                return `Module ${moduleId}`;
        }
    };

    const getModuleIcon = (moduleId: number): string => {
        const module = modules.find(m => m.id === moduleId);
        if (!module) return "";

        switch (module.module_type) {
            case 'media':
                return videoCameraIcon;
            case 'widget':
                return widgetIcon;
            case 'prompt':
                return promptIcon;
            default:
                return widgetIcon;
        }
    };



    /* Handlers for drag and drop */
    const handleDragStart = (index: number) => {
        setDraggedIndex(index);
    };

    const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
        e.preventDefault();
    };

    const handleDrop = (dropIndex: number) => {
        if (draggedIndex === null || draggedIndex === dropIndex) return;

        const targetItem = items[dropIndex];

        const updatedItems = items.map((item, index) => {
            // Dragging UP
            if (dropIndex < draggedIndex) {
                if (index === draggedIndex) {
                    return {
                        ...item,
                        position: targetItem.position
                    };
                }

                if (index >= dropIndex && index < draggedIndex) {
                    return {
                        ...item,
                        position: item.position + 1
                    };
                }
            }

            // Dragging DOWN
            if (dropIndex > draggedIndex) {
                if (index === draggedIndex) {
                    return {
                        ...item,
                        position: targetItem.position
                    };
                }

                if (index > draggedIndex && index <= dropIndex) {
                    return {
                        ...item,
                        position: item.position - 1
                    };
                }
            }

            return item;
        });

        // Reorder array visually
        const reordered = [...updatedItems];
        const [removed] = reordered.splice(draggedIndex, 1);
        reordered.splice(dropIndex, 0, removed);

        setItems(reordered);
        setDraggedIndex(null);
    };

    const handleDragEnd = () => {
        setDraggedIndex(null);
        console.log(items);
    };

    return (
        <div className="space-y-4">
            {items.map((item, index) => (
                <div 
                    key={item.id || index}
                    onDragOver={handleDragOver}
                    onDrop={() => handleDrop(index)}
                    className={`flex items-center gap-1 bg-gray-300 px-2 py-2 rounded-lg transition-opacity ${
                        draggedIndex === index ? 'opacity-50' : 'opacity-100'
                    }`}
                >
                    <div
                        draggable
                        onDragStart={() => handleDragStart(index)}
                        onDragEnd={handleDragEnd}
                        className="cursor-move flex items-center gap-1 px-2"
                    >
                        {/* Six dot drag handle */}
                        <div className="flex flex-col gap-1">
                            <div className="flex gap-1">
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                            </div>
                            <div className="flex gap-1">
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                            </div>
                            <div className="flex gap-1">
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                                <div className="w-1.5 h-1.5 bg-gray-600 rounded-full"></div>
                            </div>
                        </div>
                    </div>
                    <button className="flex-1 flex items-center gap-4 bg-slate-100 p-2 rounded-lg">
                        <img src={getModuleIcon(item.module_id)} alt="module icon" className="w-12 h-12 rounded" />
                        <div className="flex-1 flex">
                            <div className="text-gray-800 font-medium text-2xl">{getModuleTitle(item.module_id)}</div>
                        </div>
                        <span className="text-gray-700 text-2xl">&gt;</span>
                    </button>
                </div>
            ))}
        </div>
    );
}
