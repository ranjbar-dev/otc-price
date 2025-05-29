import matplotlib.pyplot as plt
import matplotlib.dates as mdates
from datetime import datetime
import numpy as np
from matplotlib.widgets import Button, Slider

# Read data from file
def load_data(filename):
    timestamps = []
    real_prices = []
    generated_prices = []
    
    with open(filename, 'r') as f:
        for line in f:
            real_price, generated_price, timestamp = map(float, line.strip().split(','))
            timestamps.append(datetime.fromtimestamp(timestamp))
            real_prices.append(real_price)
            generated_prices.append(generated_price)
    
    return timestamps, real_prices, generated_prices

# Create the plot
def create_plot():
    # Load data
    timestamps, real_prices, generated_prices = load_data('data.txt')
    
    # Create figure and axis
    fig, ax = plt.subplots(figsize=(15, 8))
    plt.subplots_adjust(bottom=0.2)  # Make room for slider
    
    # Plot the data
    line1, = ax.plot(timestamps, real_prices, label='Real Price', color='blue', alpha=0.7)
    line2, = ax.plot(timestamps, generated_prices, label='Generated Price', color='red', alpha=0.7)
    
    # Customize the plot
    ax.set_title('BTC Price Comparison')
    ax.set_xlabel('Time')
    ax.set_ylabel('Price (USDT)')
    ax.grid(True, alpha=0.3)
    ax.legend()
    
    # Format x-axis
    ax.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d %H:%M'))
    plt.xticks(rotation=45)
    
    # Add zoom slider
    ax_slider = plt.axes([0.1, 0.05, 0.65, 0.03])
    slider = Slider(
        ax=ax_slider,
        label='Zoom',
        valmin=0,
        valmax=len(timestamps),
        valinit=len(timestamps),
        valstep=1
    )
    
    # Add reset button
    ax_button = plt.axes([0.8, 0.05, 0.1, 0.03])
    reset_button = Button(ax_button, 'Reset View')
    
    def update(val):
        # Get the number of points to show
        n_points = int(slider.val)
        if n_points > 0:
            # Update the data
            line1.set_data(timestamps[-n_points:], real_prices[-n_points:])
            line2.set_data(timestamps[-n_points:], generated_prices[-n_points:])
            
            # Update the view
            ax.set_xlim(timestamps[-n_points], timestamps[-1])
            ax.set_ylim(
                min(min(real_prices[-n_points:]), min(generated_prices[-n_points:])) * 0.999,
                max(max(real_prices[-n_points:]), max(generated_prices[-n_points:])) * 1.001
            )
            fig.canvas.draw_idle()
    
    def reset(event):
        slider.reset()
        update(slider.val)
    
    # Connect the update function to the slider
    slider.on_changed(update)
    reset_button.on_clicked(reset)
    
    # Add some statistics
    diff_percent = np.array(generated_prices) - np.array(real_prices)
    diff_percent = (diff_percent / np.array(real_prices)) * 100
    
    stats_text = (
        f'Statistics:\n'
        f'Real Price - Mean: {np.mean(real_prices):.2f}, Std: {np.std(real_prices):.2f}\n'
        f'Generated Price - Mean: {np.mean(generated_prices):.2f}, Std: {np.std(generated_prices):.2f}\n'
        f'Average Difference: {np.mean(diff_percent):.2f}%'
    )
    
    plt.figtext(0.02, 0.02, stats_text, fontsize=8, 
                bbox=dict(facecolor='white', alpha=0.8))
    
    # Show the plot with standard navigation toolbar
    plt.tight_layout()
    plt.show()

if __name__ == "__main__":
    create_plot() 