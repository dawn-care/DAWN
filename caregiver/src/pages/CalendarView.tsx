/**
 * CalendarView.tsx
 * 
 * User: Caregiver
 * 
 * A simple monthly calendar UI component.
 * 
 * Responsibilities:
 * - Display a calendar grid for the current month
 * - Allow navigation between months
 * - Allow users to select a date
 * - Navigate to the agenda view for a selected date
 * 
 * Navigation:
 * - On date click → redirects to `/agenda/YYYY-MM-DD`
 * 
 * Notes:
 * - Dates are formatted manually to avoid timezone/UTC shift issues
 * - Calendar is generated dynamically based on the currentDate state
 */

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

export default function CalendarView() {
  /* ============================
     Routing
  ============================ */

  /**
   * React Router navigation hook
   * Used to navigate to agenda view on date click
   */
  const navigate = useNavigate();

  /* ============================
     State
  ============================ */

  /**
   * Tracks the currently displayed month/year
   * Defaults to today's date
   */
  const [currentDate, setCurrentDate] = useState(new Date());

  /* ============================
     Helper Functions
  ============================ */

  /**
   * Returns the number of days in a given month
   * 
   * @param date - Reference date
   * @returns number of days in the month
   */
  const getDaysInMonth = (date: Date) => {
    return new Date(date.getFullYear(), date.getMonth() + 1, 0).getDate();
  };

  /**
   * Returns the day index (0–6) of the first day of the month
   * 
   * @param date - Reference date
   * @returns index of first weekday (0 = Sunday)
   */
  const getFirstDayOfMonth = (date: Date) => {
    return new Date(date.getFullYear(), date.getMonth(), 1).getDay();
  };

  /* ============================
     Constants
  ============================ */

  /**
   * Month names for display
   */
  const monthNames = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ];

  /**
   * Day-of-week labels
   */
  const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  /* ============================
     Calendar Grid Generation
  ============================ */

  /**
   * Array representing all cells in the calendar grid
   * Includes:
   * - null values for leading empty cells
   * - numeric day values for actual dates
   */
  const days: (number | null)[] = [];
  const firstDay = getFirstDayOfMonth(currentDate);
  const daysInMonth = getDaysInMonth(currentDate);

  // Add empty cells for days before the first day of the month
  for (let i = 0; i < firstDay; i++) {
    days.push(null);
  }

  // Add the days of the month
  for (let i = 1; i <= daysInMonth; i++) {
    days.push(i);
  }

  /* ============================
     Event Handlers
  ============================ */

  /**
   * Handles clicking a calendar day
   * 
   * Converts selected day into YYYY-MM-DD format
   * and navigates to the agenda page
   * 
   * @param day - selected day of the month
   */
  const handleDateClick = (day: number) => {
    const date = new Date(currentDate.getFullYear(), currentDate.getMonth(), day);
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');

    /**
     * Format: YYYY-MM-DD without UTC shift
     * Avoids UTC conversion issues by using local date parts
     */
    const formattedDate = `${y}-${m}-${d}`; 
    navigate(`/agenda/${formattedDate}`);
  };

  /**
   * Navigate to previous month
   */
  const handlePrevMonth = () => {
    setCurrentDate(new Date(currentDate.getFullYear(), currentDate.getMonth() - 1));
  };

  /**
   * Navigate to next month
   */
  const handleNextMonth = () => {
    setCurrentDate(new Date(currentDate.getFullYear(), currentDate.getMonth() + 1));
  };

  /* ============================
     Render
  ============================ */
  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-4xl mx-auto px-8 py-12">
        {/* Page Title */}
        <h1 className="text-4xl font-light text-gray-800 mb-8 text-center">Calendar</h1>

        <div className="bg-white rounded-lg shadow-lg p-8">
          {/* ============================
              Month Navigation Header
          ============================ */}
          <div className="flex items-center justify-between mb-8">
            <button
              onClick={handlePrevMonth}
              className="text-2xl text-gray-600 hover:text-gray-900 transition"
            >
              &lt;
            </button>
            <h2 className="text-2xl font-semibold text-gray-800">
              {monthNames[currentDate.getMonth()]} {currentDate.getFullYear()}
            </h2>
            <button
              onClick={handleNextMonth}
              className="text-2xl text-gray-600 hover:text-gray-900 transition"
            >
              &gt;
            </button>
          </div>

          {/* ============================
              Day Labels (Sun–Sat)
          ============================ */}
          <div className="grid grid-cols-7 gap-2 mb-4">
            {dayNames.map((day) => (
              <div
                key={day}
                className="text-center font-semibold text-gray-600 py-2"
              >
                {day}
              </div>
            ))}
          </div>

          {/* ============================
              Calendar Grid
          ============================ */}
          <div className="grid grid-cols-7 gap-2">
            {days.map((day, index) => (
              <div
                key={index}
                className={`aspect-square flex items-center justify-center rounded-lg transition ${
                  day === null
                    ? 'bg-gray-50'
                    : 'bg-gray-100 hover:bg-blue-200 cursor-pointer text-gray-800 font-medium'
                }`}
                onClick={() => day !== null && handleDateClick(day)}
              >
                {day}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
